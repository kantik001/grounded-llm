package app

import cfgpkg "grounded_llm_server/internal/config"

// Deps holds process-wide collaborators. Prefer D over bare globals in new code.
type Deps struct {
	Config *cfgpkg.Config
	Store  *ChatStore
}

// D is populated at the start of Run.
var D Deps

// Migration aliases — same pointers as D (existing call sites).
var (
	config    *Config
	chatStore *ChatStore
)

func bindDeps(cfg *Config, store *ChatStore) {
	D.Config = cfg
	D.Store = store
	config = cfg
	chatStore = store
}
