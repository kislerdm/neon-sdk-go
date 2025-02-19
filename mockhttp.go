package sdk

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

// endpointResponseExamples denotes response mock split by
//   - the object notifier from the request path:
//     /projects - for projects
//     /project/branches - for branches
//     /project/branches/endpoints - for endpoints
//   - request REST method
var endpointResponseExamples = map[string]map[string]mockResponse{
	"/api_keys": {
		"GET": mockResponse{
			Content: `[{"created_at":"2022-11-15T20:13:35Z","created_by":{"id":"629982cc-de05-43db-ae16-28f2399c4910","image":"http://link.to.image","name":"John Smith"},"id":165432,"last_used_at":"2022-11-15T20:22:51Z","last_used_from_addr":"192.0.2.255","name":"mykey_1"},{"created_at":"2022-11-15T20:12:36Z","created_by":{"id":"629982cc-de05-43db-ae16-28f2399c4910","image":"http://link.to.image","name":"John Smith"},"id":165433,"last_used_at":"2022-11-15T20:15:04Z","last_used_from_addr":"192.0.2.255","name":"mykey_2"}]`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{"created_at":"2022-11-15T20:13:35Z","created_by":"629982cc-de05-43db-ae16-28f2399c4910","id":165434,"key":"9v1faketcjbl4sn1013keyd43n2a8qlfakeog8yvp40hx16keyjo1bpds4y2dfms3","name":"mykey"}`,
			Code:    200,
		},
	},

	"/api_keys/{key_id}": {
		"DELETE": mockResponse{
			Content: `{"created_at":"2022-11-15T20:13:35Z","created_by":"629982cc-de05-43db-ae16-28f2399c4910","id":165435,"last_used_at":"2022-11-15T20:15:04Z","last_used_from_addr":"192.0.2.255","name":"mykey","revoked":true}`,
			Code:    200,
		},
	},

	"/consumption_history/account": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/consumption_history/projects": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/organizations/{org_id}": {
		"GET": mockResponse{
			Content: `{"created_at":"2024-02-23T17:42:25Z","handle":"my-organization-my-organization-morning-bread-81040908","id":"my-organization-morning-bread-81040908","managed_by":"console","name":"my-organization","plan":"scale","updated_at":"2024-02-26T20:41:25Z"}`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/api_keys": {
		"GET": mockResponse{
			Content: `[{"created_at":"2022-11-15T20:13:35Z","created_by":{"id":"629982cc-de05-43db-ae16-28f2399c4910","image":"http://link.to.image","name":"John Smith"},"id":165432,"last_used_at":"2022-11-15T20:22:51Z","last_used_from_addr":"192.0.2.255","name":"orgkey_1"},{"created_at":"2022-11-15T20:12:36Z","created_by":{"id":"629982cc-de05-43db-ae16-28f2399c4910","image":"http://link.to.image","name":"John Smith"},"id":165433,"last_used_at":"2022-11-15T20:15:04Z","last_used_from_addr":"192.0.2.255","name":"orgkey_2"}]`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{"created_at":"2022-11-15T20:13:35Z","created_by":"629982cc-de05-43db-ae16-28f2399c4910","id":165434,"key":"9v1faketcjbl4sn1013keyd43n2a8qlfakeog8yvp40hx16keyjo1bpds4y2dfms3","name":"orgkey"}`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/api_keys/{key_id}": {
		"DELETE": mockResponse{
			Content: `{"created_at":"2022-11-15T20:13:35Z","created_by":"629982cc-de05-43db-ae16-28f2399c4910","id":165435,"last_used_at":"2022-11-15T20:15:04Z","last_used_from_addr":"192.0.2.255","name":"orgkey","revoked":true}`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/invitations": {
		"GET": mockResponse{
			Content: `{"invitations":[{"email":"invited1@email.com","id":"db8faf32-b07f-4b0f-94c8-5c288909f5d3","invited_at":"2024-02-23T17:42:25Z","invited_by":"some@email.com","org_id":"my-organization-morning-bread-81040908","role":"admin"},{"email":"invited2@email.com","id":"c52f0d22-ebd9-4708-ae44-2872cae49a83","invited_at":"2024-02-23T12:42:25Z","invited_by":"some@email.com","org_id":"my-organization-morning-bread-81040908","role":"member"}]}`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/members": {
		"GET": mockResponse{
			Content: `{"members":[{"member":{"id":"d57833f2-d308-4ede-9d2e-468d9d013d1b","joined_at":"2024-02-23T17:42:25Z","org_id":"my-organization-morning-bread-81040908","role":"admin","user_id":"b107d689-6dd2-4c9a-8b9e-0b25e457cf56"},"user":{"email":"user1@email.com"}},{"member":{"id":"5fee13ac-957b-40cd-8de0-4d494cc28e28","joined_at":"2024-02-21T16:42:25Z","org_id":"my-organization-morning-bread-81040908","role":"member","user_id":"6df052ac-ca9a-4321-8963-b6507b2d7dee"},"user":{"email":"user2@email.com"}}]}`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/members/{member_id}": {
		"DELETE": mockResponse{
			Content: `null`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"id":"d57833f2-d308-4ede-9d2e-468d9d013d1b","joined_at":"2024-02-23T17:42:25Z","org_id":"my-organization-morning-bread-81040908","role":"admin","user_id":"b107d689-6dd2-4c9a-8b9e-0b25e457cf56"}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/projects/transfer": {
		"POST": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/vpc/region/{region_id}/vpc_endpoints": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/organizations/{org_id}/vpc/region/{region_id}/vpc_endpoints/{vpc_endpoint_id}": {
		"DELETE": mockResponse{
			Content: `null`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects": {
		"GET": mockResponse{
			Content: `{"applications":{"winter-boat-259881":["vercel","github"]},"integrations":{"winter-boat-259881":["vercel","github"]},"projects":[{"active_time":100,"branch_logical_size_limit":0,"branch_logical_size_limit_bytes":10800,"cpu_used_sec":0,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","id":"shiny-wind-028834","name":"shiny-wind-028834","owner_id":"1232111","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","store_passwords":true,"updated_at":"2022-11-23T17:42:25Z"},{"active_time":100,"branch_logical_size_limit":0,"branch_logical_size_limit_bytes":10800,"cpu_used_sec":0,"created_at":"2022-11-23T17:52:25Z","creation_source":"console","id":"winter-boat-259881","name":"winter-boat-259881","org_id":"org-morning-bread-81040908","owner_id":"1232111","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","store_passwords":true,"updated_at":"2022-11-23T17:52:25Z"}]}`,
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

	"/projects/auth/create": {
		"POST": mockResponse{
			Content: `null`,
			Code:    201,
		},
	},

	"/projects/auth/keys": {
		"POST": mockResponse{
			Content: `null`,
			Code:    201,
		},
	},

	"/projects/auth/transfer_ownership": {
		"POST": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/shared": {
		"GET": mockResponse{
			Content: `{"projects":[{"active_time":100,"branch_logical_size_limit":0,"branch_logical_size_limit_bytes":10800,"cpu_used_sec":0,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","id":"shiny-wind-028834","name":"shiny-wind-028834","owner_id":"1232111","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","store_passwords":true,"updated_at":"2022-11-23T17:42:25Z"},{"active_time":100,"branch_logical_size_limit":0,"branch_logical_size_limit_bytes":10800,"cpu_used_sec":0,"created_at":"2022-11-23T17:52:25Z","creation_source":"console","id":"winter-boat-259881","name":"winter-boat-259881","owner_id":"1232111","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","store_passwords":true,"updated_at":"2022-11-23T17:52:25Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}": {
		"DELETE": mockResponse{
			Content: `{"project":{"active_time_seconds":100,"branch_logical_size_limit":0,"branch_logical_size_limit_bytes":10500,"compute_time_seconds":100,"consumption_period_end":"2023-03-01T00:00:00Z","consumption_period_start":"2023-02-01T00:00:00Z","cpu_used_sec":23004200,"created_at":"2022-11-30T18:41:29Z","creation_source":"console","data_storage_bytes_hour":1040,"data_transfer_bytes":1000000,"history_retention_seconds":604800,"id":"bold-cloud-468218","name":"bold-cloud-468218","owner_id":"1232111","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","store_passwords":true,"updated_at":"2022-11-30T18:41:29Z","written_data_bytes":100800}}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"project":{"active_time_seconds":100,"branch_logical_size_limit":0,"branch_logical_size_limit_bytes":10500,"compute_time_seconds":100,"consumption_period_end":"2023-03-01T00:00:00Z","consumption_period_start":"2023-02-01T00:00:00Z","cpu_used_sec":10,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","data_storage_bytes_hour":1040,"data_transfer_bytes":1000000,"history_retention_seconds":604800,"id":"shiny-wind-028834","name":"shiny-wind-028834","owner":{"branches_limit":10,"email":"some@email.com","name":"John Smith","subscription_type":"scale"},"owner_id":"1232111","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","store_passwords":true,"updated_at":"2022-11-23T17:42:25Z","written_data_bytes":100800}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"operations":[],"project":{"active_time_seconds":100,"branch_logical_size_limit":0,"branch_logical_size_limit_bytes":10500,"compute_time_seconds":100,"consumption_period_end":"2023-03-01T00:00:00Z","consumption_period_start":"2023-02-01T00:00:00Z","cpu_used_sec":213230,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","data_storage_bytes_hour":1040,"data_transfer_bytes":1000000,"history_retention_seconds":604800,"id":"shiny-wind-028834","name":"myproject","owner_id":"1232111","pg_version":15,"platform_id":"aws","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","store_passwords":true,"updated_at":"2022-12-04T02:39:25Z","written_data_bytes":100800}}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/auth/integration/{auth_provider}": {
		"DELETE": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/{project_id}/auth/integrations": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches": {
		"GET": mockResponse{
			Content: `{"annotations":{"br-aged-salad-637688":{"created_at":"2022-11-23T17:42:25Z","object":{"id":"br-aged-salad-637688","type":"console/branch"},"updated_at":"2022-11-23T17:42:26Z","value":{"vercel-commit-ref":"test"}}},"branches":[{"active_time_seconds":100,"compute_time_seconds":100,"cpu_used_sec":100,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","current_state":"ready","data_transfer_bytes":1000000,"default":true,"id":"br-aged-salad-637688","logical_size":28,"name":"main","project_id":"shiny-wind-028834","protected":false,"state_changed_at":"2022-11-30T20:09:48Z","updated_at":"2022-11-23T17:42:26Z","written_data_bytes":100800},{"active_time_seconds":100,"compute_time_seconds":100,"cpu_used_sec":100,"created_at":"2022-11-30T19:09:48Z","creation_source":"console","current_state":"ready","data_transfer_bytes":1000000,"default":true,"id":"br-sweet-breeze-497520","logical_size":28,"name":"dev2","parent_id":"br-aged-salad-637688","parent_lsn":"0/1DE2850","project_id":"shiny-wind-028834","protected":false,"state_changed_at":"2022-11-30T20:09:48Z","updated_at":"2022-11-30T19:09:49Z","written_data_bytes":100800},{"active_time_seconds":100,"compute_time_seconds":100,"cpu_used_sec":100,"created_at":"2022-11-30T17:36:57Z","creation_source":"console","current_state":"ready","data_transfer_bytes":1000000,"default":true,"id":"br-raspy-hill-832856","logical_size":21,"name":"dev1","parent_id":"br-aged-salad-637688","parent_lsn":"0/19623D8","project_id":"shiny-wind-028834","protected":false,"state_changed_at":"2022-11-30T20:09:48Z","updated_at":"2022-11-30T17:36:57Z","written_data_bytes":100800}],"pagination":{"next":"eyJjcmVhdGV","sort_by":"updated_at","sort_order":"desc"}}`,
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

	"/projects/{project_id}/branches/count": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}": {
		"DELETE": mockResponse{
			Content: `{"branch":{"active_time_seconds":100,"compute_time_seconds":100,"cpu_used_sec":100,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","current_state":"ready","data_transfer_bytes":1000000,"default":true,"id":"br-aged-salad-637688","logical_size":28,"name":"main","project_id":"shiny-wind-028834","protected":false,"state_changed_at":"2022-11-30T20:09:48Z","updated_at":"2022-11-23T17:42:26Z","written_data_bytes":100800},"operations":[{"action":"suspend_compute","branch_id":"br-sweet-breeze-497520","created_at":"2022-12-01T19:53:05Z","endpoint_id":"ep-soft-violet-752733","failures_count":0,"id":"b6afbc21-2990-4a76-980b-b57d8c2948f2","project_id":"shiny-wind-028834","status":"running","total_duration_ms":100,"updated_at":"2022-12-01T19:53:05Z"},{"action":"delete_timeline","branch_id":"br-sweet-breeze-497520","created_at":"2022-12-01T19:53:05Z","failures_count":0,"id":"b6afbc21-2990-4a76-980b-b57d8c2948f2","project_id":"shiny-wind-028834","status":"scheduling","total_duration_ms":100,"updated_at":"2022-12-01T19:53:05Z"}]}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"annotation":{"created_at":"2022-11-23T17:42:25Z","object":{"id":"br-aged-salad-637688","type":"console/branch"},"updated_at":"2022-11-23T17:42:26Z","value":{"vercel-commit-ref":"test"}},"branch":{"active_time_seconds":100,"compute_time_seconds":100,"cpu_used_sec":100,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","current_state":"ready","data_transfer_bytes":1000000,"default":true,"id":"br-aged-salad-637688","logical_size":28,"name":"main","project_id":"shiny-wind-028834","protected":false,"state_changed_at":"2022-11-30T20:09:48Z","updated_at":"2022-11-23T17:42:26Z","written_data_bytes":100800}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"branch":{"active_time_seconds":100,"compute_time_seconds":100,"cpu_used_sec":100,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","current_state":"ready","data_transfer_bytes":1000000,"default":true,"id":"br-icy-dream-250089","name":"mybranch","parent_id":"br-aged-salad-637688","parent_lsn":"0/1E19478","project_id":"shiny-wind-028834","protected":false,"state_changed_at":"2022-11-30T20:09:48Z","updated_at":"2022-11-23T17:42:26Z","written_data_bytes":100800},"operations":[]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/compare_schema": {
		"GET": mockResponse{
			Content: `null`,
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
			Content: `{"database":{"branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","id":876692,"name":"mydb","owner_name":"casey","updated_at":"2022-12-04T00:15:04Z"},"operations":[{"action":"apply_config","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"39426015-db00-40fa-85c5-1c7072df46d0","project_id":"shiny-wind-028834","status":"running","total_duration_ms":100,"updated_at":"2022-12-04T00:15:04Z"},{"action":"suspend_compute","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"b7483d4e-33da-4d40-b319-ac858d4d3e69","project_id":"shiny-wind-028834","status":"scheduling","total_duration_ms":100,"updated_at":"2022-12-04T00:15:04Z"}]}`,
			Code:    201,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/databases/{database_name}": {
		"DELETE": mockResponse{
			Content: `{"database":{"branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:41:46Z","id":851537,"name":"mydb","owner_name":"casey","updated_at":"2022-12-01T19:41:46Z"},"operations":[{"action":"apply_config","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:51:41Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"9ef1c2ed-dce4-43aa-bae8-78aea636bf8a","project_id":"shiny-wind-028834","status":"running","total_duration_ms":100,"updated_at":"2022-12-01T19:51:41Z"},{"action":"suspend_compute","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:51:41Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"42dafb46-f861-497b-ae89-f2bec54f4966","project_id":"shiny-wind-028834","status":"scheduling","total_duration_ms":100,"updated_at":"2022-12-01T19:51:41Z"}]}`,
			Code:    200,
		},
		"GET": mockResponse{
			Content: `{"database":{"branch_id":"br-aged-salad-637688","created_at":"2022-11-30T18:25:15Z","id":834686,"name":"main","owner_name":"casey","updated_at":"2022-11-30T18:25:15Z"}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"database":{"branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:15:04Z","id":876692,"name":"mydb","owner_name":"sally","updated_at":"2022-12-04T00:15:04Z"},"operations":[{"action":"apply_config","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:21:01Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"9ef1c2ed-dce4-43aa-bae8-78aea636bf8a","project_id":"shiny-wind-028834","status":"running","total_duration_ms":100,"updated_at":"2022-12-04T00:21:01Z"},{"action":"suspend_compute","branch_id":"br-aged-salad-637688","created_at":"2022-12-04T00:21:01Z","endpoint_id":"ep-little-smoke-851426","failures_count":0,"id":"42dafb46-f861-497b-ae89-f2bec54f4966","project_id":"shiny-wind-028834","status":"scheduling","total_duration_ms":100,"updated_at":"2022-12-04T00:21:01Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/endpoints": {
		"GET": mockResponse{
			Content: `{"endpoints":[{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","current_state":"idle","disabled":false,"host":"ep-little-smoke-851426.us-east-2.aws.neon.tech","id":"ep-little-smoke-851426","last_active":"2022-11-23T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-11-30T18:25:21Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/restore": {
		"POST": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/roles": {
		"GET": mockResponse{
			Content: `{"roles":[{"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","name":"casey","protected":false,"updated_at":"2022-11-23T17:42:25Z"},{"branch_id":"br-aged-salad-637688","created_at":"2022-10-22T17:38:21Z","name":"thomas","protected":false,"updated_at":"2022-10-22T17:38:21Z"}]}`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `{"operations":[{"action":"apply_config","branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T11:58:29Z","endpoint_id":"ep-small-pine-767857","failures_count":0,"id":"2c2be371-d5ac-4db5-8b68-79f05e8bc287","project_id":"shiny-wind-028834","status":"running","updated_at":"2022-12-03T11:58:29Z"}],"role":{"branch_id":"br-noisy-sunset-458773","created_at":"2022-12-03T11:58:29Z","name":"sally","password":"Onf1AjayKwe0","protected":false,"updated_at":"2022-12-03T11:58:29Z"}}`,
			Code:    201,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/roles/{role_name}": {
		"DELETE": mockResponse{
			Content: `{"operations":[{"action":"apply_config","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:48:11Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"db646be3-eace-4910-9f60-8150823c5cb8","project_id":"shiny-wind-028834","status":"running","total_duration_ms":100,"updated_at":"2022-12-01T19:48:11Z"},{"action":"suspend_compute","branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T19:48:11Z","endpoint_id":"ep-steep-bush-777093","failures_count":0,"id":"ab94cdad-7630-4943-a55e-5a0952d2e598","project_id":"shiny-wind-028834","status":"scheduling","total_duration_ms":100,"updated_at":"2022-12-01T19:48:11Z"}],"role":{"branch_id":"br-raspy-hill-832856","created_at":"2022-12-01T14:36:23Z","name":"thomas","protected":false,"updated_at":"2022-12-01T14:36:23Z"}}`,
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

	"/projects/{project_id}/branches/{branch_id}/roles/{role_name}/reveal_password": {
		"GET": mockResponse{
			Content: `{"password":"mypass"}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/schema": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/{project_id}/branches/{branch_id}/set_as_default": {
		"POST": mockResponse{
			Content: `{"branch":{"active_time_seconds":1,"compute_time_seconds":1,"cpu_used_sec":1,"created_at":"2022-11-23T17:42:25Z","creation_source":"console","current_state":"ready","data_transfer_bytes":100,"default":true,"id":"br-icy-dream-250089","name":"mybranch","parent_id":"br-aged-salad-637688","parent_lsn":"0/1E19478","project_id":"shiny-wind-028834","protected":false,"state_changed_at":"2022-11-30T20:09:48Z","updated_at":"2022-11-23T17:42:26Z","written_data_bytes":100},"operations":[]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/connection_uri": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/{project_id}/endpoints": {
		"GET": mockResponse{
			Content: `{"endpoints":[{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","creation_source":"console","current_state":"idle","disabled":false,"host":"ep-little-smoke-851426.us-east-2.aws.neon.tech","id":"ep-little-smoke-851426","last_active":"2022-11-23T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"suspend_timeout_seconds":10800,"type":"read_write","updated_at":"2022-11-30T18:25:21Z"},{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-11-30T17:36:57Z","creation_source":"console","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-11-30T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"suspend_timeout_seconds":10800,"type":"read_write","updated_at":"2022-11-30T18:42:58Z"},{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-sweet-breeze-497520","created_at":"2022-11-30T19:09:48Z","creation_source":"console","current_state":"idle","disabled":false,"host":"ep-soft-violet-752733.us-east-2.aws.neon.tech","id":"ep-soft-violet-752733","last_active":"2022-11-30T19:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"suspend_timeout_seconds":10800,"type":"read_write","updated_at":"2022-11-30T19:14:51Z"}]}`,
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
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-aged-salad-637688","created_at":"2022-11-23T17:42:25Z","creation_source":"console","current_state":"idle","disabled":false,"host":"ep-little-smoke-851426.us-east-2.aws.neon.tech","id":"ep-little-smoke-851426","last_active":"2022-11-23T17:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"suspend_timeout_seconds":10800,"type":"read_write","updated_at":"2022-11-30T18:25:21Z"}}`,
			Code:    200,
		},
		"PATCH": mockResponse{
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"fd11748e-3c68-458f-b9e3-66d409e3eef0","project_id":"bitter-meadow-966132","status":"running","updated_at":"2022-12-03T15:51:06Z"}]}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/endpoints/{endpoint_id}/restart": {
		"POST": mockResponse{
			Content: `{"endpoint":{"autoscaling_limit_max_cu":1,"autoscaling_limit_min_cu":1,"branch_id":"br-raspy-hill-832856","created_at":"2022-12-03T15:37:07Z","creation_source":"console","current_state":"idle","disabled":false,"host":"ep-steep-bush-777093.us-east-2.aws.neon.tech","id":"ep-steep-bush-777093","last_active":"2022-12-03T15:00:00Z","passwordless_access":true,"pooler_enabled":false,"pooler_mode":"transaction","project_id":"shiny-wind-028834","provisioner":"k8s-pod","proxy_host":"us-east-2.aws.neon.tech","region_id":"aws-us-east-2","settings":{"pg_settings":{}},"suspend_timeout_seconds":10800,"type":"read_write","updated_at":"2022-12-03T15:49:10Z"},"operations":[{"action":"suspend_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"e061087e-3c99-4856-b9c8-6b7751a253af","project_id":"bitter-meadow-966132","status":"running","total_duration_ms":100,"updated_at":"2022-12-03T15:51:06Z"},{"action":"start_compute","branch_id":"br-proud-paper-090813","created_at":"2022-12-03T15:51:06Z","endpoint_id":"ep-shrill-thunder-454069","failures_count":0,"id":"e061087e-3c99-4856-b9c8-6b7751a253af","project_id":"bitter-meadow-966132","status":"running","total_duration_ms":100,"updated_at":"2022-12-03T15:51:06Z"}]}`,
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

	"/projects/{project_id}/jwks": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `null`,
			Code:    201,
		},
	},

	"/projects/{project_id}/jwks/{jwks_id}": {
		"DELETE": mockResponse{
			Content: `null`,
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
			Content: `{"operation":{"action":"create_timeline","branch_id":"br-bitter-sound-247814","created_at":"2022-10-04T18:20:17Z","endpoint_id":"ep-dark-snowflake-942567","failures_count":0,"id":"a07f8772-1877-4da9-a939-3a3ae62d1d8d","project_id":"floral-king-961888","status":"finished","total_duration_ms":100,"updated_at":"2022-10-04T18:20:18Z"}}`,
			Code:    200,
		},
	},

	"/projects/{project_id}/vpc_endpoints": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/projects/{project_id}/vpc_endpoints/{vpc_endpoint_id}": {
		"DELETE": mockResponse{
			Content: `null`,
			Code:    200,
		},
		"POST": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/regions": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/users/me": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/users/me/organizations": {
		"GET": mockResponse{
			Content: `null`,
			Code:    200,
		},
	},

	"/users/me/projects/transfer": {
		"POST": mockResponse{
			Content: `null`,
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
func NewMockHTTPClient() MockHTTPClient {
	router := http.NewServeMux()
	u, _ := url.Parse(baseURL)
	var prefix = u.Path
	for p, httpMethodResp := range endpointResponseExamples {
		for httpMethod, resp := range httpMethodResp {
			router.HandleFunc(fmt.Sprintf("%s %s%s", httpMethod, prefix, p), func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				switch returnNotFound(r.URL.Path) {
				case true:
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte("authorization failed"))

				case false:
					w.WriteHeader(resp.Code)
					_, _ = w.Write([]byte(resp.Content))
				}

			})
		}
	}
	return MockHTTPClient{
		router: router,
	}
}

func returnNotFound(s string) bool {
	return strings.Contains(s, "notFound") ||
		strings.Contains(s, "notExist") ||
		strings.Contains(s, "notExists") ||
		strings.Contains(s, "missing")
}

type mockResponse struct {
	Content string
	Code    int
}

// MockHTTPClient defines http client to mock the SDK client.
type MockHTTPClient struct {
	router *http.ServeMux
}

func (m MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var err error
	var resp *http.Response
	if resp = authErrorResp(req); resp == nil {
		_, path := m.router.Handler(req)
		switch path != "" {
		case true:
			rec := httptest.NewRecorder()
			m.router.ServeHTTP(rec, req)
			resp = rec.Result()

		case false:
			o := Error{HTTPCode: http.StatusInternalServerError}
			o.errorResp.Message = "endpoint is not defined"
			resp = o.httpResp()
		}
	}

	return resp, err
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
