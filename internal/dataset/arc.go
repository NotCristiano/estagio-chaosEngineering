package dataset

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
)

// https://www.kaggle.com/datasets/jeromeblanchet/arc-ai2-reasoning-challenge
// https://gobyexample.com/embed-directive
// i've learnt comments have more utility xd
//
//go:embed arc-challenge.jsonl
var arcRaw []byte

type rawChoice struct {
	Text  string `json:"text"`
	Label string `json:"label"`
}

type rawQuestion struct {
	ID       string `json:"id"`
	Question struct {
		Stem    string      `json:"stem"`
		Choices []rawChoice `json:"choices"`
	} `json:"question"`
	AnswerKey string `json:"answerKey"` // A-D
}
type Question struct {
	ID      string
	Text    string
	Choices []string // Choice texts
	Labels  []string // Avaliable options (A-D)
	Answer  string   // Correct label
}

func (q Question) FormatPrompt() string {
	var sb strings.Builder
	sb.WriteString(q.Text)
	sb.WriteString("\n")
	for i, label := range q.Labels {
		if i < len(q.Choices) {
			fmt.Fprintf(&sb, "%s) %s\n", label, q.Choices[i])
		}
	}
	sb.WriteString("Answer with only the letter (A, B, C, D):")
	return sb.String()
}

func normalizeLabel(label string) string {
	switch label {
	case "1":
		return "A"
	case "2":
		return "B"
	case "3":
		return "C"
	case "4":
		return "D"
	default:
		return label
	}
}

type Arc struct {
	questions []Question
	index     int
	mutex     sync.Mutex
}

func Load() (*Arc, error) {
	var questions []Question
	decoder := json.NewDecoder(bytes.NewReader(arcRaw))
	for {
		var raw rawQuestion
		if err := decoder.Decode(&raw); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to decode ARC dataset: %w", err)
		}

		var choices, labels []string
		for _, choice := range raw.Question.Choices {
			choices = append(choices, choice.Text)
			labels = append(labels, normalizeLabel(choice.Label))
		}

		questions = append(questions, Question{
			ID:      raw.ID,
			Text:    raw.Question.Stem,
			Choices: choices,
			Labels:  labels,
			Answer:  normalizeLabel(raw.AnswerKey),
		})
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("no questions found")
	}

	fixedSeed := rand.NewSource(0401)
	random := rand.New(fixedSeed)

	random.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})

	return &Arc{questions: questions}, nil
}

func (a *Arc) Next() Question {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	question := a.questions[a.index]
	a.index = (a.index + 1) % len(a.questions)
	return question
}

func (a *Arc) Len() int {
	return len(a.questions)
}
