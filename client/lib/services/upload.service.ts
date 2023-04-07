import { grpc } from "@improbable-eng/grpc-web";

import * as uploadpb from "pb/ts/upload/v1/upload_service_pb_service";
import { GetPresignURLRequest, GetPresignURLResponse } from "pb/ts/upload/v1/upload_service_pb";

import { GRPC_GATEWAY } from '~/env';

export class UploadService {

    constructor(identityContext /* Should it be outside react context? */ = undefined) {
    }

    getUploadUrl(title: string): Promise<GetPresignURLResponse.AsObject> {
        const request = new GetPresignURLRequest();
        request.setTitle(title);

        return new Promise((resolve, reject) => {
            const client = grpc.unary(uploadpb.UploadService.GetPresignURL, {
                host: GRPC_GATEWAY, request, onEnd(response) {
                    client.close();

                    if (grpc.Code.OK != response.status)
                        reject(response.statusMessage)
                    else
                        resolve(response.message.toObject() as GetPresignURLResponse.AsObject)
                },
            })
        });
    }
}
