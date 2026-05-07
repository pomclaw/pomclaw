// Error codes matching Go pkg/protocol/errors.go

export const ErrorCodes = {
  INVALID_REQUEST: "INVALID_REQUEST",
  UNAUTHORIZED: "UNAUTHORIZED",
  NOT_FOUND: "NOT_FOUND",
  NOT_LINKED: "NOT_LINKED",
  NOT_PAIRED: "NOT_PAIRED",
  AGENT_TIMEOUT: "AGENT_TIMEOUT",
  UNAVAILABLE: "UNAVAILABLE",
  ALREADY_EXISTS: "ALREADY_EXISTS",
  RESOURCE_EXHAUSTED: "RESOURCE_EXHAUSTED",
  FAILED_PRECONDITION: "FAILED_PRECONDITION",
  INTERNAL: "INTERNAL",
} as const;

export class ApiError extends Error {
  constructor(
    public code: string,
    message: string,
    public details?: unknown,
    public retryable?: boolean,
  ) {
    super(message);
    this.name = "ApiError";
  }
}
