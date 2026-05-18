# ML Engine

This service exposes the prediction endpoints used by the Go API:

- `POST /predict/wait-time`
- `POST /predict/no-show`
- `POST /recommend/restaurants`
- `GET /health`

## Local run

```bash
pip install -r requirements.txt
uvicorn app:app --reload --port 8000
```

The current implementation includes:

- heuristic wait-time prediction
- baseline no-show probability
- LangGraph-based restaurant recommendation ranking
