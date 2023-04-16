import { grpc } from "@improbable-eng/grpc-web";
import { NodeHttpTransport } from "@improbable-eng/grpc-web-node-http-transport";
import * as audiopb from "pb/ts/audio/v1/audio_service_pb";
import { UploadError, UploadService } from "~/lib/services/upload.service"

grpc.setDefaultTransport(NodeHttpTransport());

type BucketForm = { audioBucketId: string, bucket: string }
type AudioMetaDataForm = { title: string }

export async function PUT(req: Request) {
    // TODO: Handle user identity

    const uploadService = new UploadService();

    const formData = await req.formData();
    const file = formData.get("file") as File;

    const bucketForm: BucketForm = JSON.parse(formData.get("bucket").toString()) as BucketForm;
    const audioMetaData = JSON.parse(formData.get("audio_metadata").toString()) as AudioMetaDataForm;

    try {
        const { url } = await uploadService.getUploadUrl(bucketForm.bucket, file.name);

        await fetch(url, {
            method: 'PUT',
            body: file
        });

        const audio = new audiopb.Audio();
        audio.setTitle(audioMetaData.title);

        const bucket = new audiopb.AudioBucket();
        bucket.setBucket(bucketForm.bucket);
        bucket.setAudioBucketId(bucketForm.audioBucketId);
        bucket.setOriginFile(file.name);

        await uploadService.successAudioUpload(bucket, audio);

        return new Response(JSON.stringify({ message: "OK" }), {
            status: 200
        });

    } catch (err) {
        if (err instanceof UploadError) {
            return new Response(JSON.stringify({ message: err.message }), {
                status: err.code
            });
        }

        return new Response(JSON.stringify({ message: "Something went wrong" }), {
            status: 500
        });
    }
}
