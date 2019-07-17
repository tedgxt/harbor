// {
//   "id": 8,
//   "name": "pullpushdeletehook",
//   "description": "create by guanxiatao",
//   "project_id": 2,
//   "targets": [
//       {
//           "type": "http",
//           "address": "http://10.173.32.58:9009",
//           "attachment": ""
//       }
//   ],
//   "hook_types": [
//       "pushImage",
//       "pullImage",
//       "deleteImage"
//   ],
//   "creator": "",
//   "creation_time": "2019-07-09T09:20:16.87695Z",
//   "update_time": "2019-07-15T03:33:44.964362Z",
//   "enabled": true
// }

export class Webhook {
  id: number;
  name: string;
  project_id: number;
  description: string;
  targets: Target[];
  hook_types: string[];
  creator: string;
  creation_time: Date;
  update_time: Date;
  enabled: boolean;
}

export class Target {
  type: string;
  address: string;
  attachment: string;
  secret: string;
  skip_cert_verify: boolean;

  constructor () {
    this.type = 'http';
    this.address = '';
    this.skip_cert_verify = true;
  }
}

// {
//   "hook_type": "pushImage",
//   "enabled": true,
//   "creation_time": "2019-07-09T09:20:16.87695Z",
//   "last_trigger_time": "0001-01-01T00:00:00Z"
// }
export class LastTrigger {
  enabled: boolean;
  hook_type: string;
  creation_time: Date;
  last_trigger_time: Date;
}