// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package svc

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino-ext/callbacks/apmplus"
	"github.com/cloudwego/eino-ext/libs/acl/opentelemetry"
	"github.com/cloudwego/eino/callbacks"
	"github.com/pomclaw/pomclaw/internal/callback"
	"github.com/pomclaw/pomclaw/internal/config"
	"github.com/pomclaw/pomclaw/internal/contracts"
	"github.com/pomclaw/pomclaw/internal/model"
	"github.com/pomclaw/pomclaw/internal/storage"
	"github.com/pomclaw/pomclaw/internal/tools"
	"github.com/pomclaw/pomclaw/pkg/oss"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/stores/postgres"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"go.opentelemetry.io/otel/sdk/metric"
)

type ServiceContext struct {
	Config config.Config

	// postgresql
	AgentContextFilesModel model.AgentContextFilesModel
	MemoryChunksModel      model.MemoryChunksModel
	MemoryDocumentsModel   model.MemoryDocumentsModel
	StateModel             model.StateModel
	PromptsModel           model.PromptsModel
	MetaModel              model.MetaModel
	AgentsModel            model.AgentsModel
	ProvidersModel         model.ProvidersModel
	SessionsModel          model.SessionsModel
	UsersModel             model.UsersModel
	SkillsModel            model.SkillsModel
	SkillGrantsModel       model.SkillGrantsModel
	ToolGrantsModel        model.ToolGrantsModel
	McpServersModel        model.McpServersModel
	McpAgentGrantsModel    model.McpAgentGrantsModel
	TracesModel            model.TracesModel
	SpansModel             model.SpansModel

	// manager
	SessionManager contracts.SessionManagerInterface
	MemoryStore    contracts.SqlMemoryStore
	PromptStore    contracts.PromptStoreInterface
	ToolsManager   contracts.ToolsManagerInterface

	// oss
	OssClient *oss.Client

	// redis
	Redis *redis.Redis
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

	tracesModel := model.NewTracesModel(psqlConn)
	spansModel := model.NewSpansModel(psqlConn)

	//traceExporter := callback.NewLogExporter()
	traceExporter := callback.NewPGExporter(tracesModel, spansModel)
	traceProvider := callback.NewLocalTracerProvider(traceExporter)
	meterProvider := metric.NewMeterProvider()
	opentelemetry.SetProvider(traceProvider, meterProvider)

	traceHandler, shutdown, err := apmplus.NewApmplusHandler(&apmplus.Config{
		Host:        "local",
		AppKey:      "local",
		ServiceName: c.Name,
	})
	if err != nil {
		panic(err)
	}

	proc.AddWrapUpListener(func() {
		_ = traceExporter.Shutdown(context.Background())
		_ = shutdown(context.Background())
	})

	callbacks.AppendGlobalHandlers(traceHandler)

	agentContextFilesModel := model.NewAgentContextFilesModel(psqlConn)
	promptsModel := model.NewPromptsModel(psqlConn)
	sessionsModel := model.NewSessionsModel(psqlConn)
	toolGrantsModel := model.NewToolGrantsModel(psqlConn)
	agentsModel := model.NewAgentsModel(psqlConn)
	memoryDocumentsModel := model.NewMemoryDocumentsModel(psqlConn)
	skillsModel := model.NewSkillsModel(psqlConn)
	skillGrantsModel := model.NewSkillGrantsModel(psqlConn)

	var ossClient *oss.Client
	if c.Oss.Endpoint != "" {
		var err2 error
		ossClient, err2 = oss.NewClient(oss.Config{
			Endpoint:        c.Oss.Endpoint,
			AccessKeyId:     c.Oss.AccessKeyId,
			AccessKeySecret: c.Oss.AccessKeySecret,
			BucketName:      c.Oss.BucketName,
			Directory:       c.Oss.Directory,
			OssDomain:       c.Oss.OssDomain,
		})
		if err2 != nil {
			logx.Errorf("oss client init failed: %v", err2)
		}
	}

	// redis 用于缓存 skill zip 包，未配置时为 nil，loader 将直连 OSS
	var redisClient *redis.Redis
	if c.Redis.Host != "" {
		redisClient = redis.MustNewRedis(c.Redis)
	}

	memoryStore := storage.NewMemoryStore(memoryDocumentsModel)
	promptStore := storage.NewPromptStore(promptsModel)
	sessionManager := storage.NewSessionStore(sessionsModel)
	toolsManager := tools.NewToolsManager(toolGrantsModel, agentsModel, memoryStore, memoryDocumentsModel, agentContextFilesModel, skillsModel, skillGrantsModel, ossClient, redisClient)

	return &ServiceContext{
		Config: c,

		AgentContextFilesModel: agentContextFilesModel,
		MemoryChunksModel:      model.NewMemoryChunksModel(psqlConn),
		MemoryDocumentsModel:   model.NewMemoryDocumentsModel(psqlConn),
		SessionsModel:          sessionsModel,
		PromptsModel:           promptsModel,
		ToolGrantsModel:        toolGrantsModel,
		StateModel:             model.NewStateModel(psqlConn),
		MetaModel:              model.NewMetaModel(psqlConn),
		AgentsModel:            agentsModel,
		SkillsModel:            skillsModel,
		SkillGrantsModel:       skillGrantsModel,
		ProvidersModel:         model.NewProvidersModel(psqlConn),
		UsersModel:             model.NewUsersModel(psqlConn),
		McpServersModel:        model.NewMcpServersModel(psqlConn),
		McpAgentGrantsModel:    model.NewMcpAgentGrantsModel(psqlConn),
		TracesModel:            tracesModel,
		SpansModel:             spansModel,

		SessionManager: sessionManager,
		MemoryStore:    memoryStore,
		PromptStore:    promptStore,
		ToolsManager:   toolsManager,
		OssClient:      ossClient,
		Redis:          redisClient,
	}
}
