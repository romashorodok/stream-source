'use client'
import { grpc } from "@improbable-eng/grpc-web";
import { NodeHttpTransport } from "@improbable-eng/grpc-web-node-http-transport";
import { GetUploadURL } from "pb/ts/upload/v1/upload_pb";
import { UploadAudioProcessRequest, UploadAudioStubRequest } from "pb/ts/upload/v1/upload_service_pb";
import { UploadService } from "pb/ts/upload/v1/upload_service_pb_service";
import React from "react"

const HOST = "http://localhost:10000"

grpc.setDefaultTransport(NodeHttpTransport());

function testProcess() {

    const stub = new UploadAudioStubRequest();
    stub.setStubField("test field")
    const metadata = { "x-grpc-web": "1", "x-token": "test" };

    grpc.unary(UploadService.UploadAudioStub, {
        request: stub,
        metadata: metadata,
        host: HOST,
        onEnd: (response) => {
            if (response.status === grpc.Code.OK && response.message) {
                console.log("Response received:", response.message.toObject());
            } else {
                console.log("Error:", response.statusMessage);
            }
        },
    })
}


function uploadAudioProcess() {
    const request = new UploadAudioProcessRequest();
    const uploadUrl = new GetUploadURL();
    uploadUrl.setFileName("My client side file")
    request.setGetUploadUrl(uploadUrl)

    const client = grpc.client(UploadService.UploadAudioProcess, {
        host: HOST,
    });

    client.start();
    client.send(request)
}

export default function Upload() {

    return (
        <div>
            <button onClick={uploadAudioProcess}>Send stream</button>
            <button onClick={testProcess}>Send test</button>
        </div>
    )
}
