create table spans
(
    id             serial primary key,
    trace_id       uuid                                   not null,
    parent_span_id uuid,
    agent_id       varchar(64),
    span_type      varchar(20)                            not null,
    name           text,
    start_time     timestamp with time zone default now() not null,
    end_time       timestamp with time zone,
    duration_ms    integer,
    status         varchar(20)              default 'running':: character varying,
    error          text,
    level          varchar(10)              default 'DEFAULT':: character varying,
    model          varchar(200),
    provider       varchar(50),
    input_tokens   integer,
    output_tokens  integer,
    total_cost     numeric(12, 8),
    finish_reason  varchar(50),
    model_params   jsonb,
    tool_name      varchar(200),
    tool_call_id   varchar(100),
    input_preview  text,
    output_preview text,
    metadata       jsonb,
    created_at     timestamp with time zone default now() not null
);

alter table spans
    owner to root;

create index idx_spans_trace
    on spans (trace_id, start_time);

create index idx_spans_parent
    on spans (parent_span_id) where (parent_span_id IS NOT NULL);

create index idx_spans_agent_time
    on spans (agent_id asc, created_at desc);

create index idx_spans_type
    on spans (span_type asc, created_at desc);

create index idx_spans_model
    on spans (model asc, created_at desc) where (model IS NOT NULL);

create index idx_spans_error
    on spans (status) where ((status)::text = 'error'::text);

create index idx_spans_trace_type
    on spans (trace_id, span_type);

