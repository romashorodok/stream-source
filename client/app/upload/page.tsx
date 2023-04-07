'use client'

import React from "react"
import { UploadService } from "~/lib/services/upload.service";

const instance = new UploadService()

export default function Upload() {

    React.useEffect(() => {
        instance.getUploadUrl("AudioTitle").then(console.log);
    });

    return (
        <div>
            Hello world
        </div>
    )
}
