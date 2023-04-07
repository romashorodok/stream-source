'use client'

import React, { useState } from "react"
import { NEXT_HOST } from "~/env";

function useUpload() {
    const [file, setFile] = useState<File>();
    const [message, setMessage] = useState<string>();

    function selectFile({ target: { files } }: React.ChangeEvent<HTMLInputElement>) {
        setFile(files[0]);
    }

    function submit() {
        const url = `${NEXT_HOST}/api/audio`

        const headers = {
            'Content-Type': file.type,
            'Content-Length': file.size.toString(),
        };

        fetch(url, {
            method: 'PUT',
            headers,
            body: file
        })
            .then(resp => resp.json())
            .then(setMessage);
    }

    return {
        selectFile,
        submit,
        message
    }
}

export default function Upload() {
    const { selectFile, submit, message } = useUpload()

    return (
        <div>
            <div>
                <input type="file" onChange={selectFile} />
            </div>

            <button onClick={submit}>Submit file</button>

            <h3>{JSON.stringify(message)}</h3>
        </div>
    )
}
