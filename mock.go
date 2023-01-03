package sdk

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var endpointResponseExamples = map[string]map[string]mockResponse{
	"/api_keys": {
		"GET": mockResponse{
			Content: `[{"created_at":"2022-11-15T20:13:35Z","id":165432,"last_used_at":"2022-11-15T20:22:51Z","last_used_from_addr":"192.0.2.255","name":"mykey_1"},{"created_at":"2022-11-15T20:12:36Z","id":165433,"last_used_at":"2022-11-15T20:15:04Z","last_used_from_addr":"192.0.2.255","name":"mykey_2"}]`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{"id":165434,"key":"9v1faketcjbl4sn1013keyd43n2a8qlfakeog8yvp40hx16keyjo1bpds4y2dfms3"}`,
			Code:    200,
		},
	},

	"/api_keys/{key_id}": {
		"DELETE": mockResponse{
			Content: `{"id":165435,"last_used_at":"2022-11-15T20:15:04Z","last_used_from_addr":"192.0.2.255","name":"Development environment key","revoked":true}`,
			Code:    200,
		},
	},

	"/projects": {
		"GET": mockResponse{
			Content: `{"projects":[{"created_at":"2022-11-23T17:42:25Z","id":"shiny-wind-028834","locked":false,"name":"shiny-wind-028834","pg_version":15,"platform_id":"aws","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","updated_at":"2022-11-23T17:42:25Z"},{"created_at":"2022-11-23T17:52:25Z","id":"winter-boat-259881","locked":false,"name":"winter-boat-259881","pg_version":15,"platform_id":"aws","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","updated_at":"2022-11-23T17:52:25Z"}]}`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{
		 "project": {
		   "maintenance_starts_at": "2023-01-02T20:03:02.273Z",
		   "id": "string",
		   "platform_id": "string",
		   "region_id": "string",
		   "name": "string",
		   "provisioner": "k8s-pod",
		   "default_endpoint_settings": {
		     "pg_settings": {
		       "additionalProp1": "string",
		       "additionalProp2": "string",
		       "additionalProp3": "string"
		     }
		   },
		   "pg_version": 0,
		   "created_at": "2023-01-02T20:03:02.273Z",
		   "updated_at": "2023-01-02T20:03:02.273Z",
		   "proxy_host": "string"
		 },
		 "connection_uris": [
		   {
		     "connection_uri": "string"
		   }
		 ],
		 "roles": [
		   {
		     "branch_id": "string",
		     "name": "string",
		     "password": "string",
		     "protected": true,
		     "created_at": "2023-01-02T20:03:02.273Z",
		     "updated_at": "2023-01-02T20:03:02.273Z"
		   }
		 ],
		 "databases": [
		   {
		     "id": 0,
		     "branch_id": "string",
		     "name": "string",
		     "owner_name": "string",
		     "created_at": "2023-01-02T20:03:02.273Z",
		     "updated_at": "2023-01-02T20:03:02.273Z"
		   }
		 ],
		 "operations": [
		     {
		       "id": "a07f8772-1877-4da9-a939-3a3ae62d1d8d",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "create_branch",
		       "status": "running",
		       "failures_count": 0,
		       "created_at": "2022-11-08T23:33:16Z",
		       "updated_at": "2022-11-08T23:33:20Z"
		     },
		     {
		       "id": "d8ac46eb-a757-42b1-9907-f78322ee394e",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "start_compute",
		       "status": "finished",
		       "failures_count": 0,
		       "created_at": "2022-11-15T20:02:00Z",
		       "updated_at": "2022-11-15T20:02:02Z"
		     }
		 ],
		 "branch": {
		   "id": "br-wispy-meadow-118737",
		   "project_id": "spring-example-302709",
		   "parent_id": "br-aged-salad-637688",
		   "parent_lsn": "0/1DE2850",
		   "name": "dev2",
		   "current_state": "ready",
		   "created_at": "2022-11-30T19:09:48Z",
		   "updated_at": "2022-12-01T19:53:05Z"
		 },
		 "endpoints": [
		   {
		     "host": "string",
		     "id": "string",
		     "project_id": "string",
		     "branch_id": "string",
		     "autoscaling_limit_min_cu": 0,
		     "autoscaling_limit_max_cu": 0,
		     "region_id": "string",
		     "type": "read_only",
		     "current_state": "init",
		     "pending_state": "init",
		     "settings": {
		       "pg_settings": {
		         "additionalProp1": "string",
		         "additionalProp2": "string",
		         "additionalProp3": "string"
		       }
		     },
		     "pooler_enabled": true,
		     "pooler_mode": "transaction",
		     "disabled": true,
		     "passwordless_access": true,
		     "last_active": "2023-01-02T20:03:02.273Z",
		     "created_at": "2023-01-02T20:03:02.273Z",
		     "updated_at": "2023-01-02T20:03:02.273Z",
		     "proxy_host": "string"
		   }
		 ]
		}`,
			Code: 201,
		},
	},

	"/projects/{project_id}": {
		"DELETE": mockResponse{
			Content: `{"project":{"created_at":"2022-11-30T18:41:29Z","id":"bold-cloud-468218","locked":false,"name":"bold-cloud-468218","pg_version":15,"platform_id":"aws","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","updated_at":"2022-11-30T18:41:29Z"}}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"project":{"created_at":"2022-11-23T17:42:25Z","id":"shiny-wind-028834","locked":false,"name":"shiny-wind-028834","pg_version":15,"platform_id":"aws","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","updated_at":"2022-11-23T17:42:25Z"}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"project":{"created_at":"2022-11-23T17:42:25Z","id":"shiny-wind-028834","locked":false,"name":"myproject","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","updated_at":"2022-12-04T02:39:25Z"}}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches": {
		"GET": mockResponse{
			Content: `{"branches":[{"created_at":"2022-11-23T17:42:25Z","current_state":"ready","id":"br-aged-salad-637688","logical_size":28,"name":"main","physical_size":38,"project_id":"shiny-wind-028834","updated_at":"2022-11-23T17:42:26Z"},{"created_at":"2022-11-30T19:09:48Z","current_state":"ready","id":"br-sweet-breeze-497520","logical_size":28,"name":"dev2","parent_id":"br-aged-salad-637688","parent_lsn":"0/1DE2850","project_id":"shiny-wind-028834","updated_at":"2022-11-30T19:09:49Z"},{"created_at":"2022-11-30T17:36:57Z","current_state":"ready","id":"br-raspy-hill-832856","logical_size":21,"name":"dev1","parent_id":"br-aged-salad-637688","parent_lsn":"0/19623D8","project_id":"shiny-wind-028834","updated_at":"2022-11-30T17:36:57Z"}]}`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{
		 "branch": {
		   "id": "br-wispy-meadow-118737",
		   "project_id": "spring-example-302709",
		   "parent_id": "br-aged-salad-637688",
		   "parent_lsn": "0/1DE2850",
		   "name": "dev2",
		   "current_state": "ready",
		   "created_at": "2022-11-30T19:09:48Z",
		   "updated_at": "2022-12-01T19:53:05Z"
		 },
		 "endpoints": [
		   {
		     "host": "string",
		     "id": "string",
		     "project_id": "string",
		     "branch_id": "string",
		     "autoscaling_limit_min_cu": 0,
		     "autoscaling_limit_max_cu": 0,
		     "region_id": "string",
		     "type": "read_only",
		     "current_state": "init",
		     "pending_state": "init",
		     "settings": {
		       "pg_settings": {
		         "additionalProp1": "string",
		         "additionalProp2": "string",
		         "additionalProp3": "string"
		       }
		     },
		     "pooler_enabled": true,
		     "pooler_mode": "transaction",
		     "disabled": true,
		     "passwordless_access": true,
		     "last_active": "2023-01-02T20:09:50.004Z",
		     "created_at": "2023-01-02T20:09:50.004Z",
		     "updated_at": "2023-01-02T20:09:50.004Z",
		     "proxy_host": "string"
		   }
		 ],
		 "operations": [
		     {
		       "id": "a07f8772-1877-4da9-a939-3a3ae62d1d8d",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "create_branch",
		       "status": "running",
		       "failures_count": 0,
		       "created_at": "2022-11-08T23:33:16Z",
		       "updated_at": "2022-11-08T23:33:20Z"
		     },
		     {
		       "id": "d8ac46eb-a757-42b1-9907-f78322ee394e",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "start_compute",
		       "status": "finished",
		       "failures_count": 0,
		       "created_at": "2022-11-15T20:02:00Z",
		       "updated_at": "2022-11-15T20:02:02Z"
		     }
		 ]
		}`,
			Code: 201,
		},
	},

	"/projects/{project_id}/branches/{branch_id}": {
		"DELETE": mockResponse{
			Content: `{"branch":{"created_at":"2022-11-23T17:42:25Z","current_state":"ready","id":"br-aged-salad-637688","logical_size":28,"name":"main","physical_size":38,"project_id":"shiny-wind-028834","updated_at":"2022-11-23T17:42:26Z"},"operations":[{"action":"suspend_compute","branch_id":"br-sweet-breeze-497520","created_at":"2022-12-01T19:53:05Z","endpoint_id":"ep-soft-violet-752733","failures_count":0,"id":"b6afbc21-2990-4a76-980b-b57d8c2948f2","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-01T19:53:05Z"},{"action":"delete_timeline","branch_id":"br-sweet-breeze-497520","created_at":"2022-12-01T19:53:05Z","failures_count":0,"id":"b6afbc21-2990-4a76-980b-b57d8c2948f2","project_id":"shiny-wind-028834","status":"scheduling","updated_at":"2022-12-01T19:53:05Z"}]}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"branch":{"created_at":"2022-11-23T17:42:25Z","current_state":"ready","id":"br-aged-salad-637688","logical_size":28,"name":"main","physical_size":38,"project_id":"shiny-wind-028834","updated_at":"2022-11-23T17:42:26Z"}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"branch":{"created_at":"2022-11-23T17:42:25Z","current_state":"ready","id":"br-icy-dream-250089","name":"mybranch","parent_id":"br-aged-salad-637688","parent_lsn":"0/1E19478","project_id":"shiny-wind-028834","updated_at":"2022-11-23T17:42:26Z"},"operations":[]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/databases": {
		"GET": mockResponse{
			Content: `{
		"databases": [
			{
				"id": 834686,
				"branch_id": "br-aged-salad-637688",
				"name": "main",
				"owner_name": "casey",
				"created_at": "2022-11-30T18:25:15Z",
				"updated_at": "2022-11-30T18:25:15Z"
			},
			{
				"id": 834686,
				"branch_id": "br-aged-salad-637688",
				"name": "mydb",
				"owner_name": "casey",
				"created_at": "2022-10-30T17:14:13Z",
				"updated_at": "2022-10-30T17:14:13Z"
			}
		]}`,
			Code: 200,
		},
		"POST": mockResponse{
			Content: `{"database":{"branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","id":876692,"name":"mydb","owner_name":"casey","updated_at":"2022-12-04T00:15:04Z"},"operations":[{"action":"apply_config","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"39426015-db00-40fa-85c5-1c7072df46d0","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-04T00:15:04Z"},{"action":"suspend_compute","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"b7483d4e-33da-4d40-b319-ac858d4d3e69","project_id":"shiny-wind-028834","status":"scheduling","updated_at":"2022-12-04T00:15:04Z"}]}`,
			Code:    201,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/databases/{database_name}": {
		"DELETE": mockResponse{
			Content: `{"database":{"branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:41:46Z","id":851537,"name":"mydb","owner_name":"casey","updated_at":"2022-12-01T19:41:46Z"},"operations":[{"action":"apply_config","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:51:41Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"9ef1c2ed-dce4-43aa-bae8-78aea636bf8a","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-01T19:51:41Z"},{"action":"suspend_compute","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:51:41Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"42dafb46-f861-497b-ae89-f2bec54f4966","project_id":"shiny-wind-028834","status":"scheduling","updated_at":"2022-12-01T19:51:41Z"}]}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"database":{"branch_id":"br-aged-salad-637688","created_at":"2022-11-30T18:25:15Z","id":834686,"name":"main","owner_name":"casey","updated_at":"2022-11-30T18:25:15Z"}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"database":{"branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","id":876692,"name":"mydb","owner_name":"sally","updated_at":"2022-12-04T00:15:04Z"},"operations":[{"action":"apply_config","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:21:01Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"9ef1c2ed-dce4-43aa-bae8-78aea636bf8a","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-04T00:21:01Z"},{"action":"suspend_compute","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:21:01Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"42dafb46-f861-497b-ae89-f2bec54f4966","project_id":"shiny-wind-028834","status":"scheduling","updated_at":"2022-12-04T00:21:01Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/endpoints": {
		"GET": mockResponse{
			Content: `{"endpoints":[{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","current_state":"idle","disabled":false,"host":"ep-little-smoke-851426.us-east-2.aws.neon.tech","id":"ep-little-smoke-851426","last_active":"2022-11-23T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-11-30T18:25:21Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/roles": {
		"GET": mockResponse{
			Content: `{"roles":[{"branch_id":null,"created_at":"2022-11-23T17:42:25Z","name":"casey","protected":false,"updated_at":"2022-11-23T17:42:25Z"},{"branch_id":null,"created_at":"2022-10-22T17:38:21Z","name":"thomas","protected":false,"updated_at":"2022-10-22T17:38:21Z"}]}`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{"operations":[{"action":"apply_config","branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T11:58:29Z","endpoint_id":"ep-small-pine-767857","failures_count":0,"id":"2c2be371-d5ac-4db5-8b68-79f05e8bc287","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-03T11:58:29Z"}],"role":{"branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T11:58:29Z","name":"sally","password":"Onf1AjayKwe0","protected":false,"updated_at":"2022-12-03T11:58:29Z"}}`,
			Code:    201,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/roles/{role_name}": {
		"DELETE": mockResponse{
			Content: `{"operations":[{"action":"apply_config","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:48:11Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"db646be3-eace-4910-9f60-8150823c5cb8","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-01T19:48:11Z"},{"action":"suspend_compute","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:48:11Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"ab94cdad-7630-4943-a55e-5a0952d2e598","project_id":"shiny-wind-028834","status":"scheduling","updated_at":"2022-12-01T19:48:11Z"}],"role":{"branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T14:36:23Z","name":"thomas","protected":false,"updated_at":"2022-12-01T14:36:23Z"}}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"role":{"branch_id":"br-noisy-sunset-458773","created_at":"2022-11-23T17:42:25Z","name":"casey","protected":false,"updated_at":"2022-11-23T17:42:25Z"}}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/roles/{role_name}/reset_password": {
		"POST": mockResponse{
			Content: `{"operations":[{"action":"apply_config","branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T12:58:18Z","endpoint_id":"ep-small-pine-767857","failures_count":0,"id":"6bef07a0-ebca-40cd-9100-7324036cfff2","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-03T12:58:18Z"},{"action":"suspend_compute","branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T12:58:18Z","endpoint_id":"ep-small-pine-767857","failures_count":0,"id":"16b5bfca-4697-4194-a338-d2cdc9aca2af","project_id":"shiny-wind-028834","status":"scheduling","updated_at":"2022-12-03T12:58:18Z"}],"role":{"branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T12:39:39Z","name":"sally","password":"ClfD0aVuK3eK","protected":false,"updated_at":"2022-12-03T12:58:18Z"}}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/endpoints": {
		"GET": mockResponse{
			Content: `{"endpoints":[{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","current_state":"idle","disabled":false,"host":"ep-little-smoke-851426.us-east-2.aws.neon.tech","id":"ep-little-smoke-851426","last_active":"2022-11-23T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-11-30T18:25:21Z"},{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-11-30T17:36:57Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-11-30T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-11-30T18:42:58Z"},{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-sweet-breeze-497520","created_at":"2022-11-30T19:09:48Z","current_state":"idle","disabled":false,"host":"ep-soft-violet-752733.us-east-2.aws.neon.tech","id":"ep-soft-violet-752733","last_active":"2022-11-30T19:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-11-30T19:14:51Z"}]}`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{
		 "endpoint": {
		   "autoscaling_limit_max_cu": 1,
		   "autoscaling_limit_min_cu": 1,
		   "branch_id": "br-proud-paper-090813",
		   "created_at": "2022-12-03T15:37:07Z",
		   "current_state": "init",
		   "disabled": false,
		   "host": "ep-shrill-thunder-454069.us-east-2.aws.neon.tech",
		   "id": "ep-shrill-thunder-454069",
		   "passwordless_access": true,
		   "pending_state": "active",
		   "pooler_enabled": false,
		   "pooler_mode": "transaction",
		   "project_id": "bitter-meadow-966132",
		   "proxy_host": "us-east-2.aws.neon.tech",
		   "region_id": "aws-us-east-2",
		   "settings": {
		     "pg_settings": {}
		   },
		   "type": "read_write",
		   "updated_at": "2022-12-03T15:37:07Z"
		 },
		 "operations": [{
		   "action": "start_compute",
		   "branch_id": "br-proud-paper-090813",
		   "created_at": "2022-12-03T15:37:07Z",
		   "endpoint_id": "ep-shrill-thunder-454069",
		   "failures_count": 0,
		   "id": "874f8bfe-f51d-4c61-85af-a29bea73e0e2",
		   "project_id": "bitter-meadow-966132",
		   "status": "running",
		   "updated_at": "2022-12-03T15:37:07Z"
		 }]
		}`,
			Code: 201,
		},
	},

	"/projects/{project_id}/endpoints/{endpoint_id}": {
		"DELETE": mockResponse{
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"fd11748e-3c68-458f-b9e3-66d409e3eef0","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","current_state":"idle","disabled":false,"host":"ep-little-smoke-851426.us-east-2.aws.neon.tech","id":"ep-little-smoke-851426","last_active":"2022-11-23T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-11-30T18:25:21Z"}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"fd11748e-3c68-458f-b9e3-66d409e3eef0","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/endpoints/{endpoint_id}/start": {
		"POST": mockResponse{
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"start_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"e061087e-3c99-4856-b9c8-6b7751a253af","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/endpoints/{endpoint_id}/suspend": {
		"POST": mockResponse{
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"e061087e-3c99-4856-b9c8-6b7751a253af","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/operations": {
		"GET": mockResponse{
			Content: `{
		 "operations": [
		     {
		       "id": "a07f8772-1877-4da9-a939-3a3ae62d1d8d",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "create_branch",
		       "status": "running",
		       "failures_count": 0,
		       "created_at": "2022-11-08T23:33:16Z",
		       "updated_at": "2022-11-08T23:33:20Z"
		     },
		     {
		       "id": "d8ac46eb-a757-42b1-9907-f78322ee394e",
		       "project_id": "spring-example-302709",
		       "branch_id": "br-wispy-meadow-118737",
		       "endpoint_id": "ep-silent-smoke-806639",
		       "action": "start_compute",
		       "status": "finished",
		       "failures_count": 0,
		       "created_at": "2022-11-15T20:02:00Z",
		       "updated_at": "2022-11-15T20:02:02Z"
		     }
		 ],
		 "pagination": {
		   "cursor": "string"
		 }
		}`,
			Code: 200,
		},
	},

	"/projects/{project_id}/operations/{operation_id}": {
		"GET": mockResponse{
			Content: `{"operation":{"action":"create_timeline","branch_id":"br-bitter-sound-247814","created_at":"2022-10-04T18:20:17Z","endpoint_id":"ep-dark-snowflake-942567","failures_count":0,"id":"a07f8772-1877-4da9-a939-3a3ae62d1d8d","project_id":"floral-king-961888","status":"finished","updated_at":"2022-10-04T18:20:18Z"}}`,
			Code:    200,
		},
	},
}

// NewMockHTTPClient initiates a mock fo the HTTP client required for the SDK client.
// Mock client return the response as per API spec, except for the errors: 404 and 401 status codes are covered only.
// - 401 is returned when the string `invalidApiKey` is used as the API key;
// - 404 is returned if either of the following:
//   - the string value `notFound` is used as the string argument, e.g. projectID
//   - a negative int/float value is used as the int/float argument, e.g. database ID
func NewMockHTTPClient() HTTPClient {
	u, _ := url.Parse(baseURL)
	return mockHTTPClient{
		endpoints:   endpointResponseExamples,
		routePrefix: u.Path,
	}
}

type mockResponse struct {
	Content string
	Code    int
}

// mockHTTPClient defines http client to mock the SDK client.
type mockHTTPClient struct {
	// endpoints denotes response mock split by
	// - the object notifier from the request path:
	// 		/projects - for projects
	// 		/project/branches - for branches
	// 		/project/branches/endpoints - for endpoints
	// - request REST method
	endpoints map[string]map[string]mockResponse

	routePrefix string
}

func (m mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if r := authErrorResp(req); r != nil {
		return r, nil
	}

	p := parsePath(strings.TrimPrefix(req.URL.Path, m.routePrefix))

	endpoint, ok := m.endpoints[p.path]
	if !ok {
		o := Error{HTTPCode: http.StatusBadRequest}
		o.errorResp.Message = "unknown endpoint"
		return o.httpResp(), nil
	}

	resp, ok := endpoint[req.Method]
	if !ok {
		o := Error{HTTPCode: http.StatusMethodNotAllowed}
		o.errorResp.Message = "method not allowed"
		return o.httpResp(), nil
	}

	if p.objNotFound {
		o := Error{HTTPCode: http.StatusNotFound}
		o.errorResp.Message = "object not found"
		return o.httpResp(), nil
	}

	return &http.Response{
		Status:        "OK",
		StatusCode:    resp.Code,
		Body:          io.NopCloser(strings.NewReader(resp.Content)),
		ContentLength: int64(len(resp.Content)),
		Request:       req,
	}, nil
}

type objPath struct {
	path        string
	objNotFound bool
}

func parsePath(s string) objPath {
	s = strings.TrimPrefix(s, "/")
	o := ""
	var notFoundReq bool
	splt := strings.Split(s, "/")
	for i, el := range splt {
		if len(el) == 0 {
			continue
		}

		if i%2 == 0 {
			o += "/" + el
			continue
		}

		if el == "notFound" || el == "notExist" || el == "notExists" || el == "missing" {
			notFoundReq = true
		}
		if v, err := strconv.ParseInt(el, 10, 64); nil == err && v < 0 {
			notFoundReq = true
		}
		if v, err := strconv.ParseFloat(el, 64); nil == err && v < 0 {
			notFoundReq = true
		}

		switch v := splt[i-1]; v {
		case "projects", "endpoints", "operations":
			o += "/{" + v[:len(v)-1] + "_id}"
		case "databases", "roles":
			o += "/{" + v[:len(v)-1] + "_name}"
		case "api_keys":
			o += "/{key_id}"
		case "branches":
			o += "/{branch_id}"
		}

	}
	return objPath{
		path:        o,
		objNotFound: notFoundReq,
	}
}

func authErrorResp(req *http.Request) *http.Response {
	token := req.Header.Get("Authorization")
	if token == "Bearer invalidApiKey" {
		o := Error{HTTPCode: http.StatusForbidden}
		o.Message = "authorization failed"
		return o.httpResp()
	}
	return nil
}
