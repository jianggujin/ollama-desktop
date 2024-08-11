// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {ollama} from '../models';
import {app} from '../models';

export function Delete(arg1:ollama.DeleteRequest):Promise<void>;

export function Envs():Promise<Array<app.OllamaEnvVar>>;

export function Heartbeat():Promise<void>;

export function LibraryOnline(arg1:ollama.LibraryRequest):Promise<Array<ollama.ModelInfo>>;

export function List():Promise<ollama.ListResponse>;

export function ListRunning():Promise<ollama.ProcessResponse>;

export function ModelInfoOnline(arg1:string):Promise<ollama.ModelInfoResponse>;

export function Pull(arg1:string,arg2:ollama.PullRequest):Promise<void>;

export function SearchOnline(arg1:ollama.SearchRequest):Promise<ollama.SearchResponse>;

export function Show(arg1:ollama.ShowRequest):Promise<ollama.ShowResponse>;

export function Start():Promise<void>;

export function Version():Promise<string>;
