/*

This is a k6 test script that imports the xk6-kafka and
tests Kafka with a 200 JSON messages per iteration.

 */
import {
    Writer,
    Reader,
    Connection,
    SchemaRegistry,
    CODEC_SNAPPY,
    SCHEMA_TYPE_STRING,
    SCHEMA_TYPE_JSON,
} from "k6/x/kafka"; // import kafka extension


const brokers = ["broker:29092"];
const topic = "flodesk.account.subscriber_event";

const writer = new Writer({
    brokers: brokers,
    topic: topic,
    autoCreateTopic: true,
    compression: CODEC_SNAPPY,
});
const reader = new Reader({
    brokers: brokers,
    topic: topic,
});
const connection = new Connection({
    address: brokers[0],
});
const schemaRegistry = new SchemaRegistry();

if (__VU == 0) {
    connection.createTopic({
        topic: topic,
        configEntries: [
            {
                configName: "compression.type",
                configValue: CODEC_SNAPPY,
            },
        ],
    });
}

export const options = {
    thresholds: {
        // Base thresholds to see if the writer or reader is working
        kafka_writer_error_count: ["count == 0"],
        kafka_reader_error_count: ["count == 0"],
    },
};

const preDefinedXIDs = [
    "c84gk48qvh2l9t5dzp3f",
    "b17t5g83ch2m6a4hjb9e",
    "f72nj40exk7s1p8wrm0t",
    "d35sk27lvh4c2j5azr8u",
    "e90fj34oxk6r2p7vhc1y",
    "a24gk57qvh1n8t5dzp6f",
    "g83t5j72ch4m6a5hjb9e",
    "h19nj50dxk3s1p8wrm2t",
    "i64sk36lvh2c5j6azr4u",
    "j20fj44oxk9r2p7vhc3y",
    "k54gk38qvh6l7t5dzp2f",
    "l32t5g92ch8m4a3hjb7e",
    "m62nj45exk2s1p9wrm5t",
    "n47sk25lvh5c2j4azr1u",
    "o21fj49oxk7r3p6vhc8y",
    "p73gk52qvh9l6t4dzp1f",
    "q18t5j81ch3m7a2hjb4e",
    "r91nj54dxk5s1p9wrm3t",
    "s53sk29lvh8c4j3azr6u",
    "t29fj42oxk6r4p8vhc5y"
];

const genRandomEventName = () => {
    const preDefinedEventNames = ["subscriber.created", "subscriber.subscribed", "subscriber.unsubscribed"];
    return preDefinedEventNames[Math.floor(Math.random() * preDefinedEventNames.length)];
}

const genRandomXIDFromPredefined = () => {
    return preDefinedXIDs[Math.floor(Math.random() * preDefinedXIDs.length)];
}

export default function () {
    console.log("execute my producer function")
    const kafkaBatchMsgNumber = 100
    const messages = [];
    for (let index = 0; index < kafkaBatchMsgNumber; index++) {
        const msg = {
            key: schemaRegistry.serialize({
                data: genRandomXIDFromPredefined(),
                schemaType: SCHEMA_TYPE_STRING,
            }),
            value: schemaRegistry.serialize({
                data: {
                    event_name: genRandomEventName(),
                    event_time: new Date().toISOString(),
                    subscriber: {
                        id: `sub${Math.floor(Math.random() * 1000000)}`,
                        status: "active",
                        email: `user${Math.floor(Math.random() * 1000)}@example.com`,
                        source: "website",
                        first_name: "John",
                        last_name: "Doe",
                        segments: [
                            {
                                id: `seg${Math.floor(Math.random() * 100)}`,
                                name: "Segment Name"
                            }
                        ],
                        custom_fields: {
                            property1: "custom_value1",
                            property2: "custom_value2"
                        },
                        optin_ip: "192.168.1.1",
                        optin_timestamp: new Date().toISOString(),
                        created_at: new Date().toISOString()
                    },
                    segment: {
                        id: `seg${Math.floor(Math.random() * 100)}`,
                        name: "Segment Name"
                    },
                    webhook_id: genRandomXIDFromPredefined()
                },
                schemaType: SCHEMA_TYPE_JSON
            }),
            headers: {
                mykey: "myvalue"
            },
            offset: index,
            partition: 0,
            time: new Date() // Will be converted to timestamp automatically
        };
        messages.push(msg);
    }
    writer.produce({ messages: messages });
}


export function teardown(data) {
    writer.close();
    reader.close();
    connection.close();
}