export namespace app {
	
	export class ChatMessage {
	    id: string;
	    sessionId: string;
	    role: string;
	    content: string;
	    success: boolean;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.success = source["success"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ConversationRequest {
	    sessionId: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ConversationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.content = source["content"];
	    }
	}
	export class ConversationResponse {
	    id: string;
	    sessionId: string;
	    content: string;
	    // Go type: time
	    createdAt: any;
	
	    static createFrom(source: any = {}) {
	        return new ConversationResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.content = source["content"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProgressBar {
	    name: string;
	    percentage: number;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new ProgressBar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.percentage = source["percentage"];
	        this.status = source["status"];
	    }
	}
	export class DownloadItem {
	    model: string;
	    insecure?: boolean;
	    bars: ProgressBar[];
	
	    static createFrom(source: any = {}) {
	        return new DownloadItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.insecure = source["insecure"];
	        this.bars = this.convertValues(source["bars"], ProgressBar);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OllamaEnvVar {
	
	
	    static createFrom(source: any = {}) {
	        return new OllamaEnvVar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class SessionHistoryMessageRequest {
	    sessionId: string;
	    nextMarker: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionHistoryMessageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sessionId = source["sessionId"];
	        this.nextMarker = source["nextMarker"];
	    }
	}
	export class SessionModel {
	    id: string;
	    sessionName: string;
	    modelName: string;
	    messageHistoryCount: number;
	    keepAlive?: string;
	    systemMessage?: string;
	    options?: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new SessionModel(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionName = source["sessionName"];
	        this.modelName = source["modelName"];
	        this.messageHistoryCount = source["messageHistoryCount"];
	        this.keepAlive = source["keepAlive"];
	        this.systemMessage = source["systemMessage"];
	        this.options = source["options"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace ollama {
	
	export class DeleteRequest {
	    model: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new DeleteRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.name = source["name"];
	    }
	}
	export class LibraryRequest {
	    q: string;
	    sort: string;
	
	    static createFrom(source: any = {}) {
	        return new LibraryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.q = source["q"];
	        this.sort = source["sort"];
	    }
	}
	export class ModelDetails {
	    parent_model: string;
	    format: string;
	    family: string;
	    families: string[];
	    parameter_size: string;
	    quantization_level: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelDetails(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parent_model = source["parent_model"];
	        this.format = source["format"];
	        this.family = source["family"];
	        this.families = source["families"];
	        this.parameter_size = source["parameter_size"];
	        this.quantization_level = source["quantization_level"];
	    }
	}
	export class ListModelResponse {
	    name: string;
	    model: string;
	    // Go type: time
	    modified_at: any;
	    size: number;
	    digest: string;
	    details?: ModelDetails;
	
	    static createFrom(source: any = {}) {
	        return new ListModelResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.model = source["model"];
	        this.modified_at = this.convertValues(source["modified_at"], null);
	        this.size = source["size"];
	        this.digest = source["digest"];
	        this.details = this.convertValues(source["details"], ModelDetails);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ListResponse {
	    models: ListModelResponse[];
	
	    static createFrom(source: any = {}) {
	        return new ListResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.models = this.convertValues(source["models"], ListModelResponse);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ToolCallFunction {
	    name: string;
	    arguments: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new ToolCallFunction(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.arguments = source["arguments"];
	    }
	}
	export class ToolCall {
	    function: ToolCallFunction;
	
	    static createFrom(source: any = {}) {
	        return new ToolCall(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.function = this.convertValues(source["function"], ToolCallFunction);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Message {
	    role: string;
	    content: string;
	    images?: number[][];
	    tool_calls?: ToolCall[];
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.images = source["images"];
	        this.tool_calls = this.convertValues(source["tool_calls"], ToolCall);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class ModelInfo {
	    name: string;
	    archive: boolean;
	    description: string;
	    pullCount: string;
	    tags: string[];
	    tagCount: number;
	    updateTime: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.archive = source["archive"];
	        this.description = source["description"];
	        this.pullCount = source["pullCount"];
	        this.tags = source["tags"];
	        this.tagCount = source["tagCount"];
	        this.updateTime = source["updateTime"];
	    }
	}
	export class ModelMeta {
	    name: string;
	    content: string;
	    unit: string;
	    href: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelMeta(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.content = source["content"];
	        this.unit = source["unit"];
	        this.href = source["href"];
	    }
	}
	export class ModelTag {
	    name: string;
	    latest: boolean;
	    id: string;
	    size: string;
	    updateTime: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelTag(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.latest = source["latest"];
	        this.id = source["id"];
	        this.size = source["size"];
	        this.updateTime = source["updateTime"];
	    }
	}
	export class ModelInfoResponse {
	    model?: ModelInfo;
	    tags: ModelTag[];
	    metas: ModelMeta[];
	    readme: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelInfoResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = this.convertValues(source["model"], ModelInfo);
	        this.tags = this.convertValues(source["tags"], ModelTag);
	        this.metas = this.convertValues(source["metas"], ModelMeta);
	        this.readme = source["readme"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProcessModelResponse {
	    name: string;
	    model: string;
	    size: number;
	    digest: string;
	    details?: ModelDetails;
	    // Go type: time
	    expires_at: any;
	    size_vram: number;
	
	    static createFrom(source: any = {}) {
	        return new ProcessModelResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.model = source["model"];
	        this.size = source["size"];
	        this.digest = source["digest"];
	        this.details = this.convertValues(source["details"], ModelDetails);
	        this.expires_at = this.convertValues(source["expires_at"], null);
	        this.size_vram = source["size_vram"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProcessResponse {
	    models: ProcessModelResponse[];
	
	    static createFrom(source: any = {}) {
	        return new ProcessResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.models = this.convertValues(source["models"], ProcessModelResponse);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PullRequest {
	    model: string;
	    insecure?: boolean;
	    username: string;
	    password: string;
	    stream?: boolean;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new PullRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.insecure = source["insecure"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.stream = source["stream"];
	        this.name = source["name"];
	    }
	}
	export class SearchRequest {
	    q: string;
	    p: number;
	    c: string;
	
	    static createFrom(source: any = {}) {
	        return new SearchRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.q = source["q"];
	        this.p = source["p"];
	        this.c = source["c"];
	    }
	}
	export class SearchResponse {
	    query: string;
	    page: number;
	    pageCount: number;
	    items: ModelInfo[];
	
	    static createFrom(source: any = {}) {
	        return new SearchResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.query = source["query"];
	        this.page = source["page"];
	        this.pageCount = source["pageCount"];
	        this.items = this.convertValues(source["items"], ModelInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ShowRequest {
	    model: string;
	    system: string;
	    template: string;
	    verbose: boolean;
	    options: {[key: string]: any};
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new ShowRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.system = source["system"];
	        this.template = source["template"];
	        this.verbose = source["verbose"];
	        this.options = source["options"];
	        this.name = source["name"];
	    }
	}
	export class ShowResponse {
	    license?: string;
	    modelfile?: string;
	    parameters?: string;
	    template?: string;
	    system?: string;
	    details?: ModelDetails;
	    messages?: Message[];
	    model_info?: {[key: string]: any};
	    projector_info?: {[key: string]: any};
	    // Go type: time
	    modified_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new ShowResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.license = source["license"];
	        this.modelfile = source["modelfile"];
	        this.parameters = source["parameters"];
	        this.template = source["template"];
	        this.system = source["system"];
	        this.details = this.convertValues(source["details"], ModelDetails);
	        this.messages = this.convertValues(source["messages"], Message);
	        this.model_info = source["model_info"];
	        this.projector_info = source["projector_info"];
	        this.modified_at = this.convertValues(source["modified_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

