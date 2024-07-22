export namespace app {
	
	export class OllamaEnvVar {
	
	
	    static createFrom(source: any = {}) {
	        return new OllamaEnvVar(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}

}

export namespace ollama {
	
	export class Duration {
	
	
	    static createFrom(source: any = {}) {
	        return new Duration(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	
	    }
	}
	export class Message {
	    role: string;
	    content: string;
	    images?: number[][];
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	        this.images = source["images"];
	    }
	}
	export class ChatRequest {
	    model: string;
	    messages: Message[];
	    stream?: boolean;
	    format: string;
	    // Go type: Duration
	    keep_alive?: any;
	    options: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new ChatRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.messages = this.convertValues(source["messages"], Message);
	        this.stream = source["stream"];
	        this.format = source["format"];
	        this.keep_alive = this.convertValues(source["keep_alive"], null);
	        this.options = source["options"];
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
	export class EmbedRequest {
	    model: string;
	    input: any;
	    // Go type: Duration
	    keep_alive?: any;
	    truncate?: boolean;
	    options: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new EmbedRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.input = source["input"];
	        this.keep_alive = this.convertValues(source["keep_alive"], null);
	        this.truncate = source["truncate"];
	        this.options = source["options"];
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
	export class EmbedResponse {
	    model: string;
	    embeddings: number[][];
	
	    static createFrom(source: any = {}) {
	        return new EmbedResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.embeddings = source["embeddings"];
	    }
	}
	export class EmbeddingRequest {
	    model: string;
	    prompt: string;
	    // Go type: Duration
	    keep_alive?: any;
	    options: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new EmbeddingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.prompt = source["prompt"];
	        this.keep_alive = this.convertValues(source["keep_alive"], null);
	        this.options = source["options"];
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
	export class EmbeddingResponse {
	    embedding: number[];
	
	    static createFrom(source: any = {}) {
	        return new EmbeddingResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.embedding = source["embedding"];
	    }
	}
	export class GenerateRequest {
	    model: string;
	    prompt: string;
	    system: string;
	    template: string;
	    context?: number[];
	    stream?: boolean;
	    raw?: boolean;
	    format: string;
	    // Go type: Duration
	    keep_alive?: any;
	    images?: number[][];
	    options: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new GenerateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.prompt = source["prompt"];
	        this.system = source["system"];
	        this.template = source["template"];
	        this.context = source["context"];
	        this.stream = source["stream"];
	        this.raw = source["raw"];
	        this.format = source["format"];
	        this.keep_alive = this.convertValues(source["keep_alive"], null);
	        this.images = source["images"];
	        this.options = source["options"];
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

