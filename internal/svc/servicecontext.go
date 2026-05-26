// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"fmt"
	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/svc/toolsmanager"
	"github.com/pomclaw/pomclaw/pkg/contracts"
	"github.com/pomclaw/pomclaw/pkg/storage"
	"github.com/zeromicro/go-zero/core/stores/postgres"
)

type ServiceContext struct {
	Config config.Config

	// postgresql
	DailyNotesModel  model.DailyNotesModel
	MemoriesModel    model.MemoriesModel
	StateModel       model.StateModel
	PromptsModel     model.PromptsModel
	MetaModel        model.MetaModel
	AgentsModel      model.AgentsModel
	ProvidersModel   model.ProvidersModel
	SessionsModel    model.SessionsModel
	UsersModel       model.UsersModel
	SkillsModel      model.SkillsModel
	SkillGrantsModel model.SkillGrantsModel
	ToolGrantsModel  model.ToolGrantsModel
	TracesModel      model.TracesModel

	// manager
	SessionManager contracts.SessionManagerInterface
	MemoryStore    contracts.SqlMemoryStore
	PromptStore    contracts.PromptStoreInterface
	ToolsManager   contracts.ToolsManagerInterface
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

	dailyNotesModel := model.NewDailyNotesModel(psqlConn)
	memoriesModel := model.NewMemoriesModel(psqlConn)
	promptsModel := model.NewPromptsModel(psqlConn)
	sessionsModel := model.NewSessionsModel(psqlConn)
	toolGrantsModel := model.NewToolGrantsModel(psqlConn)
	agentsModel := model.NewAgentsModel(psqlConn)

	memoryStore := storage.NewMemoryStore(memoriesModel, dailyNotesModel)
	promptStore := storage.NewPromptStore(promptsModel)
	sessionManager := storage.NewSessionStore(sessionsModel)

	toolsManager := toolsmanager.NewToolsManager(memoryStore, toolGrantsModel, agentsModel)

	return &ServiceContext{
		Config: c,

		DailyNotesModel:  dailyNotesModel,
		MemoriesModel:    memoriesModel,
		SessionsModel:    sessionsModel,
		PromptsModel:     promptsModel,
		ToolGrantsModel:  toolGrantsModel,
		StateModel:       model.NewStateModel(psqlConn),
		MetaModel:        model.NewMetaModel(psqlConn),
		AgentsModel:      agentsModel,
		SkillsModel:      model.NewSkillsModel(psqlConn),
		SkillGrantsModel: model.NewSkillGrantsModel(psqlConn),
		ProvidersModel:   model.NewProvidersModel(psqlConn),
		UsersModel:       model.NewUsersModel(psqlConn),
		TracesModel:      model.NewTracesModel(psqlConn),

		SessionManager: sessionManager,
		MemoryStore:    memoryStore,
		PromptStore:    promptStore,
		ToolsManager:   toolsManager,
	}
}
