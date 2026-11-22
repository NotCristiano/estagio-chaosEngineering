# Fault Injection and Resilience Testing for Microservices using Chaos Engineering

This document presents the original academic assignment in a structured form and explains how its scope evolved during implementation. The original two-page brief is preserved in [`docs/original-project-brief.pdf`](docs/original-project-brief.pdf).

## Context and project goal

Design and implement a fault-injection and resilience-testing system for microservices using Chaos Engineering principles. Introduce failures deliberately in a controlled test environment, observe the system's response, and evaluate its ability to tolerate disruption and recover.

## Expected outcomes

1. A functional prototype for fault injection and resilience testing.
2. An evaluation of how effectively the prototype simulates failures and supports resilience analysis.
3. Recommendations for improving the system's usability and reliability.

These are the assignment's intended outcomes. They should not be read as evidence that every proposed test or evaluation was completed.

## Technologies proposed in the original brief

| Area | Proposed technologies |
| --- | --- |
| Programming | Python and Go |
| Chaos experiments | [Chaos Toolkit](https://chaostoolkit.org/) and [Simmy](https://github.com/Polly-Contrib/Simmy) as candidate tools |
| Containerisation | Docker |
| Orchestration | Kubernetes |
| Monitoring | Prometheus and Grafana |

## Methodology

1. **Literature review:** study Chaos Engineering, fault injection and resilience testing in microservices.
2. **System design and implementation:** choose suitable tools, build the testing platform, and prepare a service-based test application.
3. **Evaluation:** conduct controlled experiments, collect and analyse measurements, and identify improvements to the system.

## Original 14-week plan

| Period | Focus | Planned activities |
| --- | --- | --- |
| Weeks 1–2 | Literature review and planning | Review relevant research; identify goals; define the scope, timeline, milestones and project plan. |
| Weeks 3–6 | Design and implementation | Design the testing platform; select chaos tools and container/orchestration technologies; implement fault injection; develop the testbed application. |
| Weeks 7–10 | Evaluation and testing | Evaluate fault simulation and resilience; investigate performance and scalability; collect data for different failure scenarios; identify improvements. |
| Weeks 11–14 | Analysis and reporting | Analyse results, identify patterns, draw conclusions, and write the final report with recommendations. |

This table records the initial plan, rather than an assertion that the final implementation followed each phase or completed a scalability study.

## Implemented scope

The internship report describes an individual project developed at the University of Évora in 2025/2026 by Cristiano Cascarrinho, supervised by Rui Oliveira. The implemented testbed uses a Go client named **Poker** and an **Ollama** service deployed on a local **k3s** cluster.

| Initial proposal | Implemented approach and rationale |
| --- | --- |
| Kubernetes orchestration | k3s provided a lighter local Kubernetes environment. |
| Python and Go | Go implements Poker; Python tooling is used by Chaos Toolkit, with dependencies managed through `uv`. |
| Chaos Toolkit and possible use of Simmy | Chaos Toolkit orchestrates experiments. Chaos Mesh provides Kubernetes fault injection; Simmy was not adopted because its .NET focus did not match this Go/Ollama testbed. |
| A microservice-based test application | Poker sends generation requests to Ollama and observes the dependency's behaviour. |
| Prometheus and Grafana monitoring | Application metrics and five custom dashboards cover ARC evaluation, application internals, containers, hardware and Kubernetes. |
| Resilience evaluation | The report discusses CPU stress, network delay, pod deletion, storage-pressure eviction and an invalid model configuration. Additional scenario definitions exist in the repository. |

The application and target service run in `chaos-production`, monitoring in `monitoring`, and Chaos Mesh controllers in `chaos-testing`. The namespace name `chaos-production` is a lab naming convention; it does not mean these experiments were performed against a production system.

## Evaluation approach used in the project

The experiments compare normal operation with behaviour during a fault and after recovery:

1. Allow the environment to stabilise and collect baseline measurements.
2. Define the scenario and its expected steady state.
3. Introduce the selected fault for the configured period.
4. Observe latency, success/failure outcomes, retries, resource use and pod state.
5. Restore the normal configuration and observe recovery.
6. Compare Chaos Toolkit journals/reports with Prometheus and Grafana timelines.

Poker also sends periodic ARC-Challenge questions to assess whether the faults affect answer correctness. A fixed dataset shuffle makes question ordering reproducible. This supplements the availability and performance measurements; it is not a general evaluation of model quality.

## Reported observations

- CPU stress increased response latency and CPU temperature.
- A network delay of 150 ms with 50% correlation increased latency without the same CPU/temperature effects.
- Pod deletion and storage-pressure eviction caused temporary dependency failures followed by automatic pod replacement and restored communication.
- An invalid model configuration left Poker running while requests could not complete successfully.
- ARC-Challenge accuracy did not show a consistent decline under the tested faults. Latency, availability and timeouts were more informative in this setup.

These findings are qualitative observations from the reported runs. They do not establish uninterrupted availability, a universal recovery time, or a general relationship between resource pressure and model accuracy. Historical screenshots can use different settings from the current experiment definitions.

## Deliverables and evidence

| Deliverable | Location |
| --- | --- |
| Poker application | [`cmd/poker/`](cmd/poker/) and [`internal/`](internal/) |
| Container and Kubernetes setup | [`docker/`](docker/) and [`k8s/`](k8s/) |
| Dashboards | [`k8s/dashboards/`](k8s/dashboards/) |
| Fault experiments and report commands | [`chaos/`](chaos/) |
| Architecture, screenshots and operating instructions | [`README.md`](README.md) |
| Original assignment | [`docs/original-project-brief.pdf`](docs/original-project-brief.pdf) |

The [internship report](https://docs.google.com/document/d/1SJ2ge7uZ9S4-QSTJBCGWu5XQYScrM951/edit), especially Sections 4–5, documents the implementation choices, experimental process and observations. The [project presentation](https://docs.google.com/presentation/d/1QucPlHPwDxAyIRzCfNlD1s7xqI2FMh4GpxLYUXq_VHE/edit) provides the architecture and screenshots used in the README. Source-document access depends on the existing Google Drive sharing permissions.

For current commands and configuration, use the repository files and README. For the original academic requirements, use the preserved brief and the first sections of this document.
