# Local model weights (optional)

Placeholder for **large binary weights** that must not be committed to git (ONNX, PyTorch, GGUF, etc.).

- Tracked: only this README and `.gitkeep`.
- Ignored: `*.pth`, `*.onnx`, `*.bin`, `*.gguf`, `*.safetensors` (see repo `.gitignore`).
- Embeddings for RAG use Hugging Face cache (`.cache/huggingface`), not this folder.
- Optional toxicity / custom classifiers may mount `./models:/models:ro` in Compose — create files locally when needed.

Do not put secrets or customer data here.
