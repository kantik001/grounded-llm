/** HTTP client for Grounded LLM REST API (Go server). */

export type ChatMessage = {
  role?: string;
  content?: string;
  [key: string]: unknown;
};

export type MessageResult = {
  sessionId: string;
  domainId: string;
  tenantId: string;
  messages: ChatMessage[];
  success: boolean;
  error?: string;
  raw: Record<string, unknown>;
};

export type GroundedClientOptions = {
  baseUrl?: string;
  apiKey?: string;
  tenantId?: string;
  locale?: string;
  apiPrefix?: string;
  timeoutMs?: number;
  fetch?: typeof fetch;
};

export class GroundedAPIError extends Error {
  statusCode?: number;
  payload?: unknown;
  constructor(message: string, statusCode?: number, payload?: unknown) {
    super(message);
    this.name = "GroundedAPIError";
    this.statusCode = statusCode;
    this.payload = payload;
  }
}

export class GroundedAuthError extends GroundedAPIError {
  constructor(message: string, statusCode?: number, payload?: unknown) {
    super(message, statusCode, payload);
    this.name = "GroundedAuthError";
  }
}

export class GroundedClient {
  readonly baseUrl: string;
  readonly apiKey?: string;
  readonly tenantId: string;
  readonly locale: string;
  readonly apiPrefix: string;
  readonly timeoutMs: number;
  private readonly fetchImpl: typeof fetch;

  constructor(opts: GroundedClientOptions = {}) {
    this.baseUrl = (opts.baseUrl ?? "http://localhost:8080").replace(/\/$/, "");
    this.apiKey = opts.apiKey;
    this.tenantId = opts.tenantId ?? "default";
    this.locale = opts.locale ?? "en";
    this.apiPrefix = (opts.apiPrefix ?? "/api/v1").replace(/\/$/, "");
    this.timeoutMs = opts.timeoutMs ?? 120_000;
    this.fetchImpl = opts.fetch ?? fetch;
  }

  private headers(jsonBody: boolean): HeadersInit {
    const h: Record<string, string> = {
      "X-Tenant-ID": this.tenantId,
      "X-Locale": this.locale,
    };
    if (this.apiKey) h["X-API-Key"] = this.apiKey;
    if (jsonBody) h["Content-Type"] = "application/json; charset=utf-8";
    return h;
  }

  private url(path: string, versioned = true): string {
    let p = path.startsWith("/") ? path : `/${path}`;
    if (versioned && this.apiPrefix && !p.startsWith(this.apiPrefix)) {
      p = `${this.apiPrefix}${p}`;
    }
    return `${this.baseUrl}${p}`;
  }

  private async request(
    method: string,
    path: string,
    opts: {
      versioned?: boolean;
      params?: Record<string, string>;
      jsonBody?: Record<string, unknown>;
    } = {},
  ): Promise<Response> {
    const versioned = opts.versioned !== false;
    const u = new URL(this.url(path, versioned));
    if (opts.params) {
      for (const [k, v] of Object.entries(opts.params)) u.searchParams.set(k, v);
    }
    const controller = new AbortController();
    const timer = setTimeout(() => controller.abort(), this.timeoutMs);
    try {
      const resp = await this.fetchImpl(u.toString(), {
        method,
        headers: this.headers(opts.jsonBody !== undefined),
        body: opts.jsonBody !== undefined ? JSON.stringify(opts.jsonBody) : undefined,
        signal: controller.signal,
      });
      if (resp.status === 401 || resp.status === 403) {
        throw new GroundedAuthError(`Authentication failed: HTTP ${resp.status}`, resp.status);
      }
      return resp;
    } finally {
      clearTimeout(timer);
    }
  }

  private async parseJson(resp: Response): Promise<Record<string, unknown>> {
    let data: Record<string, unknown>;
    try {
      data = (await resp.json()) as Record<string, unknown>;
    } catch {
      throw new GroundedAPIError(`Invalid JSON response: HTTP ${resp.status}`, resp.status);
    }
    if (resp.status >= 400 || data.success === false) {
      throw new GroundedAPIError(
        String(data.error ?? `HTTP ${resp.status}`),
        resp.status,
        data,
      );
    }
    return data;
  }

  async health(): Promise<Record<string, unknown>> {
    return this.parseJson(await this.request("GET", "/health", { versioned: false }));
  }

  async listDomains(): Promise<Record<string, unknown>> {
    return this.parseJson(
      await this.request("GET", "/api/domains", {
        versioned: false,
        params: { locale: this.locale },
      }),
    );
  }

  async createSession(domainId = "default"): Promise<string> {
    const data = await this.parseJson(
      await this.request("POST", "/session", { jsonBody: { domain_id: domainId } }),
    );
    const sessionId = data.session_id;
    if (!sessionId) throw new GroundedAPIError("Missing session_id in response", undefined, data);
    return String(sessionId);
  }

  async history(sessionId: string): Promise<ChatMessage[]> {
    const data = await this.parseJson(
      await this.request("GET", "/history", { params: { session_id: sessionId } }),
    );
    return (data.messages as ChatMessage[]) ?? [];
  }

  async sendMessage(
    text: string,
    opts: { sessionId: string; domainId?: string },
  ): Promise<MessageResult> {
    const domainId = opts.domainId ?? "default";
    const data = await this.parseJson(
      await this.request("POST", "/message", {
        jsonBody: {
          session_id: opts.sessionId,
          domain_id: domainId,
          text,
        },
      }),
    );
    return {
      sessionId: String(data.session_id ?? opts.sessionId),
      domainId: String(data.domain_id ?? domainId),
      tenantId: String(data.tenant_id ?? this.tenantId),
      messages: (data.messages as ChatMessage[]) ?? [],
      success: true,
      raw: data,
    };
  }

  async chat(
    text: string,
    opts: { domainId?: string; sessionId?: string } = {},
  ): Promise<MessageResult> {
    const domainId = opts.domainId ?? "default";
    const sessionId = opts.sessionId ?? (await this.createSession(domainId));
    return this.sendMessage(text, { sessionId, domainId });
  }

  lastAssistant(result: MessageResult): ChatMessage | undefined {
    for (let i = result.messages.length - 1; i >= 0; i--) {
      if (result.messages[i]?.role === "assistant") return result.messages[i];
    }
    return undefined;
  }
}
