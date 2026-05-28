create table traces
(
    id                  serial primary key,
    parent_trace_id     uuid,
    agent_id            varchar(64),
    user_id             varchar(255),
    session_key         text,
    run_id              text,
    start_time          timestamp with time zone default now() not null,
    end_time            timestamp with time zone,
    duration_ms         integer,
    name                text,
    channel             varchar(50),
    input_preview       text,
    output_preview      text,
    total_input_tokens  integer                  default 0,
    total_output_tokens integer                  default 0,
    total_cost          numeric(12, 6)           default 0,
    span_count          integer                  default 0,
    llm_call_count      integer                  default 0,
    tool_call_count     integer                  default 0,
    status              varchar(20)              default 'running':: character varying,
    error               text,
    metadata            jsonb,
    tags                text[],
    created_at          timestamp with time zone default now() not null
);

alter table traces
    owner to root;

create index idx_traces_agent_time
    on traces (agent_id asc, created_at desc);

create index idx_traces_user_time
    on traces (user_id asc, created_at desc) where (user_id IS NOT NULL);

create index idx_traces_session
    on traces (session_key asc, created_at desc) where (session_key IS NOT NULL);

create index idx_traces_status
    on traces (status) where ((status)::text = 'error'::text);

create index idx_traces_parent
    on traces (parent_trace_id) where (parent_trace_id IS NOT NULL);

create index idx_traces_quota
    on traces (user_id asc, created_at desc) where ((parent_trace_id IS NULL) AND (user_id IS NOT NULL));

create index idx_traces_start_root
    on traces (start_time desc) where (parent_trace_id IS NULL);
