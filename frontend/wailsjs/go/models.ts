export namespace config {

	export class ToolSpec {
	    id: string;
	    name: string;
	    repo: string;
	    binary: string;
	    asset_pattern: string;
	    checksums_pattern: string;
	    platform_map?: Record<string, string>;
	    os?: string;
	    arch?: string;
	    version_cmd?: string[];
	    version_regex?: string;
	    version_transform?: string;
	    install_dir?: string;

	    static createFrom(source: any = {}) {
	        return new ToolSpec(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.repo = source["repo"];
	        this.binary = source["binary"];
	        this.asset_pattern = source["asset_pattern"];
	        this.checksums_pattern = source["checksums_pattern"];
	        this.platform_map = source["platform_map"];
	        this.os = source["os"];
	        this.arch = source["arch"];
	        this.version_cmd = source["version_cmd"];
	        this.version_regex = source["version_regex"];
	        this.version_transform = source["version_transform"];
	        this.install_dir = source["install_dir"];
	    }
	}

}

export namespace main {

	export class ReleaseNote {
	    tag_name: string;
	    name: string;
	    body: string;
	    published_at: string;

	    static createFrom(source: any = {}) {
	        return new ReleaseNote(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tag_name = source["tag_name"];
	        this.name = source["name"];
	        this.body = source["body"];
	        this.published_at = source["published_at"];
	    }
	}
	export class ToolExplanation {
	    summary: string;
	    summary_en?: string;
	    releases: ReleaseNote[];

	    static createFrom(source: any = {}) {
	        return new ToolExplanation(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.summary = source["summary"];
	        this.summary_en = source["summary_en"];
	        this.releases = this.convertValues(source["releases"], ReleaseNote);
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

export namespace tool {

	export class InstallResult {
	    tool_id: string;
	    version: string;
	    bin_path: string;
	    operation: string;

	    static createFrom(source: any = {}) {
	        return new InstallResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tool_id = source["tool_id"];
	        this.version = source["version"];
	        this.bin_path = source["bin_path"];
	        this.operation = source["operation"];
	    }
	}
	export class ToolStatus {
	    spec: config.ToolSpec;
	    installed: boolean;
	    installed_version: string;
	    installed_from: string;
	    latest_version: string;
	    update_available: boolean;
	    error: string;

	    static createFrom(source: any = {}) {
	        return new ToolStatus(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.spec = this.convertValues(source["spec"], config.ToolSpec);
	        this.installed = source["installed"];
	        this.installed_version = source["installed_version"];
	        this.installed_from = source["installed_from"];
	        this.latest_version = source["latest_version"];
	        this.update_available = source["update_available"];
	        this.error = source["error"];
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
