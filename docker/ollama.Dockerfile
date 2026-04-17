FROM ollama/ollama:latest

# We pull locally the model, wait a bit and then we commit the image
RUN nohup bash -c "ollama serve &" && sleep 5 && ollama pull llama3.2