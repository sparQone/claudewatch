export namespace main {
	
	export class SessionInfo {
	    id: string;
	    projectName: string;
	    projectPath: string;
	    usedTokens: number;
	    freeTokens: number;
	    percentage: number;
	    lastUpdated: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.projectName = source["projectName"];
	        this.projectPath = source["projectPath"];
	        this.usedTokens = source["usedTokens"];
	        this.freeTokens = source["freeTokens"];
	        this.percentage = source["percentage"];
	        this.lastUpdated = source["lastUpdated"];
	    }
	}

}

