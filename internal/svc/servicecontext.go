// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"fmt"
	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/storage"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/postgres"
)

type ServiceContext struct {
	Config config.Config

	// postgresql
	DailyNotesModel  model.DailyNotesModel
	MemoriesModel    model.MemoriesModel
	StateModel       model.StateModel
	SkillGrantsModel model.SkillGrantsModel
	PromptsModel     model.PromptsModel
	MetaModel        model.MetaModel
	AgentsModel      model.AgentsModel
	SkillsModel      model.SkillsModel
	ProvidersModel   model.ProvidersModel
	SessionsModel    model.SessionsModel
	UsersModel       model.UsersModel

	// manager
	SessionManager contracts.SessionManagerInterface
	MemoryStore    contracts.SqlMemoryStore
	PromptStore    contracts.PromptStoreInterface
}

func NewServiceContext(c config.Config) *ServiceContext {

	// BuildConnStr constructs the PostgreSQL connection string.
	psqlConn := postgres.New(fmt.Sprintf(
		"host=%s port=%d database=%s user=%s password=%s sslmode=%s",
		c.Postgres.Host,
		c.Postgres.Port,
		c.Postgres.Database,
		c.Postgres.User,
		c.Postgres.Password,
		c.Postgres.SSLMode,
	))

	if c.LangfuseConfig.Enabled {

		cbh, _ := langfuse.NewLangfuseHandler(&langfuse.Config{
			Host:      c.LangfuseConfig.Host,
			PublicKey: c.LangfuseConfig.PublicKey,
			SecretKey: c.LangfuseConfig.SecretKey,
			Name:      c.LangfuseConfig.Name,
			Public:    c.LangfuseConfig.Public,
			Release:   c.LangfuseConfig.Release,
			UserID:    c.LangfuseConfig.UserID,
			Tags:      c.LangfuseConfig.Tags,
		})
		if cbh == nil {
			panic("langfuse failed")
		}

		callbacks.AppendGlobalHandlers(cbh)
	}

	// Connect using factory
	conn, err := storage.NewConnectionManager(&c)
	if err != nil {
		panic(err)
	}

	embSvc, err := storage.NewEmbeddingService(&c, conn.DB())
	if err != nil {
		logx.Error("agent", "Failed to create embedding service", map[string]interface{}{"error": err.Error()})
		panic(err)
	}

	memoryStore := storage.NewMemoryStore(&c, conn.DB(), embSvc)
	promptStore := storage.NewPromptStore(&c, conn.DB())
	sessionManager := storage.NewSessionStore(&c, conn.DB())

	//a, err := agent.NewAgentLoop(&c, memoryStore, promptStore, sessionManager)

	return &ServiceContext{
		Config: c,

		DailyNotesModel:  model.NewDailyNotesModel(psqlConn),
		MemoriesModel:    model.NewMemoriesModel(psqlConn),
		StateModel:       model.NewStateModel(psqlConn),
		SkillGrantsModel: model.NewSkillGrantsModel(psqlConn),
		PromptsModel:     model.NewPromptsModel(psqlConn),
		MetaModel:        model.NewMetaModel(psqlConn),
		AgentsModel:      model.NewAgentsModel(psqlConn),
		SkillsModel:      model.NewSkillsModel(psqlConn),
		ProvidersModel:   model.NewProvidersModel(psqlConn),
		SessionsModel:    model.NewSessionsModel(psqlConn),
		UsersModel:       model.NewUsersModel(psqlConn),

		SessionManager: sessionManager,
		MemoryStore:    memoryStore,
		PromptStore:    promptStore,
	}
}
