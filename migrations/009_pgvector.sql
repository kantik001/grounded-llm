-- Enable pgvector for VECTOR_STORE=pgvector (requires Postgres image with pgvector extension).
CREATE EXTENSION IF NOT EXISTS vector;
