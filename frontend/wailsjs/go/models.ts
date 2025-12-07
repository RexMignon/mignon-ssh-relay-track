export namespace config {
	
	export class IConfigLinkGroup {
	    id: string;
	    name: string;
	    local_host: string;
	    remote_host: string;
	    remote_port: number;
	    local_port: number;
	    notes: string;
	    is_penetrate: boolean;
	    is_open: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IConfigLinkGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.local_host = source["local_host"];
	        this.remote_host = source["remote_host"];
	        this.remote_port = source["remote_port"];
	        this.local_port = source["local_port"];
	        this.notes = source["notes"];
	        this.is_penetrate = source["is_penetrate"];
	        this.is_open = source["is_open"];
	    }
	}
	export class IConfigGroup {
	    id: string;
	    username: string;
	    password: string;
	    server_name: string;
	    server_host: string;
	    server_port: number;
	    link_group: IConfigLinkGroup[];
	    is_open: boolean;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new IConfigGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.server_name = source["server_name"];
	        this.server_host = source["server_host"];
	        this.server_port = source["server_port"];
	        this.link_group = this.convertValues(source["link_group"], IConfigLinkGroup);
	        this.is_open = source["is_open"];
	        this.notes = source["notes"];
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
	export class IConfig {
	    config: IConfigGroup[];
	    is_dark: boolean;
	    is_english: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], IConfigGroup);
	        this.is_dark = source["is_dark"];
	        this.is_english = source["is_english"];
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

