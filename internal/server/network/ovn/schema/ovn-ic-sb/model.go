// Code generated by "libovsdb.modelgen"
// DO NOT EDIT.

package ovsmodel

import (
	"encoding/json"

	"github.com/ovn-org/libovsdb/model"
	"github.com/ovn-org/libovsdb/ovsdb"
)

// FullDatabaseModel returns the DatabaseModel object to be used in libovsdb
func FullDatabaseModel() (model.ClientDBModel, error) {
	return model.NewClientDBModel("OVN_IC_Southbound", map[string]model.Model{
		"Availability_Zone": &AvailabilityZone{},
		"Connection":        &Connection{},
		"Datapath_Binding":  &DatapathBinding{},
		"Encap":             &Encap{},
		"Gateway":           &Gateway{},
		"IC_SB_Global":      &ICSBGlobal{},
		"Port_Binding":      &PortBinding{},
		"Route":             &Route{},
		"SSL":               &SSL{},
	})
}

var schema = `{
  "name": "OVN_IC_Southbound",
  "version": "1.1.0",
  "tables": {
    "Availability_Zone": {
      "columns": {
        "name": {
          "type": "string"
        }
      },
      "indexes": [
        [
          "name"
        ]
      ],
      "isRoot": true
    },
    "Connection": {
      "columns": {
        "external_ids": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "inactivity_probe": {
          "type": {
            "key": {
              "type": "integer"
            },
            "min": 0,
            "max": 1
          }
        },
        "is_connected": {
          "type": "boolean",
          "ephemeral": true
        },
        "max_backoff": {
          "type": {
            "key": {
              "type": "integer",
              "minInteger": 1000
            },
            "min": 0,
            "max": 1
          }
        },
        "other_config": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "status": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          },
          "ephemeral": true
        },
        "target": {
          "type": "string"
        }
      },
      "indexes": [
        [
          "target"
        ]
      ]
    },
    "Datapath_Binding": {
      "columns": {
        "external_ids": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "transit_switch": {
          "type": "string"
        },
        "tunnel_key": {
          "type": {
            "key": {
              "type": "integer",
              "minInteger": 1,
              "maxInteger": 16777215
            }
          }
        }
      },
      "indexes": [
        [
          "tunnel_key"
        ]
      ],
      "isRoot": true
    },
    "Encap": {
      "columns": {
        "gateway_name": {
          "type": "string"
        },
        "ip": {
          "type": "string"
        },
        "options": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "type": {
          "type": {
            "key": {
              "type": "string",
              "enum": [
                "set",
                [
                  "geneve",
                  "stt",
                  "vxlan"
                ]
              ]
            }
          }
        }
      },
      "indexes": [
        [
          "type",
          "ip"
        ]
      ]
    },
    "Gateway": {
      "columns": {
        "availability_zone": {
          "type": {
            "key": {
              "type": "uuid",
              "refTable": "Availability_Zone"
            }
          }
        },
        "encaps": {
          "type": {
            "key": {
              "type": "uuid",
              "refTable": "Encap"
            },
            "min": 1,
            "max": "unlimited"
          }
        },
        "external_ids": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "hostname": {
          "type": "string"
        },
        "name": {
          "type": "string"
        }
      },
      "indexes": [
        [
          "name"
        ]
      ],
      "isRoot": true
    },
    "IC_SB_Global": {
      "columns": {
        "connections": {
          "type": {
            "key": {
              "type": "uuid",
              "refTable": "Connection"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "external_ids": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "options": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "ssl": {
          "type": {
            "key": {
              "type": "uuid",
              "refTable": "SSL"
            },
            "min": 0,
            "max": 1
          }
        }
      },
      "isRoot": true
    },
    "Port_Binding": {
      "columns": {
        "address": {
          "type": "string"
        },
        "availability_zone": {
          "type": {
            "key": {
              "type": "uuid",
              "refTable": "Availability_Zone"
            }
          }
        },
        "encap": {
          "type": {
            "key": {
              "type": "uuid",
              "refTable": "Encap",
              "refType": "weak"
            },
            "min": 0,
            "max": 1
          }
        },
        "external_ids": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "gateway": {
          "type": "string"
        },
        "logical_port": {
          "type": "string"
        },
        "transit_switch": {
          "type": "string"
        },
        "tunnel_key": {
          "type": {
            "key": {
              "type": "integer",
              "minInteger": 1,
              "maxInteger": 32767
            }
          }
        }
      },
      "indexes": [
        [
          "transit_switch",
          "tunnel_key"
        ],
        [
          "logical_port"
        ]
      ],
      "isRoot": true
    },
    "Route": {
      "columns": {
        "availability_zone": {
          "type": {
            "key": {
              "type": "uuid",
              "refTable": "Availability_Zone"
            }
          }
        },
        "external_ids": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "ip_prefix": {
          "type": "string"
        },
        "nexthop": {
          "type": "string"
        },
        "origin": {
          "type": {
            "key": {
              "type": "string",
              "enum": [
                "set",
                [
                  "connected",
                  "static"
                ]
              ]
            }
          }
        },
        "route_table": {
          "type": "string"
        },
        "transit_switch": {
          "type": "string"
        }
      },
      "isRoot": true
    },
    "SSL": {
      "columns": {
        "bootstrap_ca_cert": {
          "type": "boolean"
        },
        "ca_cert": {
          "type": "string"
        },
        "certificate": {
          "type": "string"
        },
        "external_ids": {
          "type": {
            "key": {
              "type": "string"
            },
            "value": {
              "type": "string"
            },
            "min": 0,
            "max": "unlimited"
          }
        },
        "private_key": {
          "type": "string"
        },
        "ssl_ciphers": {
          "type": "string"
        },
        "ssl_protocols": {
          "type": "string"
        }
      }
    }
  }
}`

func Schema() ovsdb.DatabaseSchema {
	var s ovsdb.DatabaseSchema
	err := json.Unmarshal([]byte(schema), &s)
	if err != nil {
		panic(err)
	}
	return s
}
