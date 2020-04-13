/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

db.createUser(
    {
        user: "kie",
        pwd: "123",
        roles:[
            {
                role: "readWrite",
                db:   "kie"
            }
        ]
    }
);
db.createCollection("counter");
db.counter.insertOne( { name:"revision_counter",count: 1,domain:"default" } );
db.createCollection( "kv", {
    validator: { $jsonSchema: {
            bsonType: "object",
            required: [ "key","domain","project","id","value","create_revision","update_revision","value_type","label_format" ],
            properties: {
                key: {
                    bsonType: "string",
                },
                value_type: {
                    enum: [ "text","string","yaml", "json", "properties", "ini" ]
                },
                labels: {
                    bsonType: "object"
                },
                create_time: {
                    bsonType: "long",
                },
                update_time: {
                    bsonType: "long",
                },
                status: {
                    bsonType: "string",
                },
            }
        } }
} );
db.createCollection("kv_revision");
db.createCollection( "view", {
    validator: { $jsonSchema: {
            bsonType: "object",
            required: [ "id","domain","project","display","criteria" ],
            properties: {
                id: {
                    bsonType: "string",
                },
                domain: {
                    bsonType: "string"
                },
                project: {
                    bsonType: "string"
                },
                criteria: {
                    bsonType: "string"
                }
            }
        } }
} );

db.createCollection( "polling_detail", {
    capped: true,
    max: 100,
    validator: { $jsonSchema: {
            bsonType: "object",
            required: [ "id","session_id","domain","url_path" ],
            properties: {
                id: {
                    bsonType: "string",
                },
                domain: {
                    bsonType: "string",
                },
                params: {
                    bsonType: "object"
                },
                ip: {
                    bsonType: "string"
                },
                user_agent: {
                    bsonType: "string"
                },
                response_body: {
                    bsonType: "object"
                },
                response_header: {
                    bsonType: "object"
                },
                response_code: {
                    bsonType: "number"
                }
            }
        } }
} );

//index
db.kv.createIndex({"id": 1}, { unique: true } );
db.kv.createIndex({key: 1, label_format: 1,domain:1,project:1},{ unique: true });
db.kv_revision.createIndex( { "delete_time": 1 }, { expireAfterSeconds: 7 * 24 * 3600 } );
db.label.createIndex({"id": 1}, { unique: true } );
db.label.createIndex({format: 1,domain:1,project:1},{ unique: true });
db.polling_detail.createIndex({"id": 1}, { unique: true } );
db.polling_detail.createIndex({session_id:1,domain:1}, { unique: true } );
db.view.createIndex({"id": 1}, { unique: true } );
db.view.createIndex({display:1,domain:1,project:1},{ unique: true });
//db config
db.setProfilingLevel(1, {slowms: 80, sampleRate: 1} );