import { grpc } from "@improbable-eng/grpc-web";
import { NodeHttpTransport } from "@improbable-eng/grpc-web-node-http-transport";
import { UploadService } from "~/lib/services/upload.service"

grpc.setDefaultTransport(NodeHttpTransport());

export async function PUT(req: Request) {
    // TODO: Handle user identity

    const uploadService = new UploadService();
    const file = await req.blob();

    try {

        const { url } = await uploadService.getUploadUrl("STUB");

        await fetch(url, {
            method: 'PUT',
            body: file
        });

        console.log("Successfull upload on ", url);

        return new Response(JSON.stringify({ message: "OK" }), {
            status: 200
        });

    } catch (err) {
        console.error(err);

        return new Response(JSON.stringify({ message: "Something went wrong" }), {
            status: 500
        });
    }
}
