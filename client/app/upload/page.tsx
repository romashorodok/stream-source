'use client'

import { AudioBucket } from "pb/ts/audio/v1/audio_service_pb";
import React, { useCallback, useMemo, useState } from "react"
import { NEXT_HOST } from "~/env";
import { AuidoService } from "~/lib/services/audio.service";


function useUpload() {
    const [file, setFile] = useState<File>();
    const [message, setMessage] = useState<string>();

    function selectFile({ target: { files } }: React.ChangeEvent<HTMLInputElement>) {
        setFile(files[0]);
    }

    function submit(audioMetaData: AudioMetaDataForm, bucket: AudioBucket.AsObject) {
        const url = `${NEXT_HOST}/api/audio`

        const formData = new FormData();
        formData.set("bucket", JSON.stringify(bucket))
        formData.set("file", file)
        formData.set("audio_metadata", JSON.stringify(audioMetaData))

        console.log(bucket)

        fetch(url, {
            method: 'PUT',
            body: formData
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

const audioService = new AuidoService()

type AudioMetaDataForm = {
    title: string
}

export default function Upload() {
    const { selectFile, submit, message } = useUpload()

    const [audioBucket, setAudioBucket] = useState<AudioBucket.AsObject>()

    const [formState, setFormState] = React.useReducer<React.Reducer<AudioMetaDataForm, Partial<AudioMetaDataForm>>>(
        (state, next) => ({ ...state, ...next }),
        { title: null }
    );

    useMemo(async () => {
        const bucket = await audioService.createBucket()
        setAudioBucket(bucket.audioBucket)
        console.log(bucket)
    }, [])

    function onChange(e: React.FormEvent) {
        if (e.target instanceof HTMLInputElement && e.target?.name != "file") {
            const { name, value } = e.target;
            setFormState({ [name]: value });
        }
    }

    return (
        <form onChange={onChange}>
            <div>
                <input name="title" type="text" />
                <input name="file" type="file" onChange={selectFile} />
            </div>

            <button type="button" onClick={() => submit(formState, audioBucket)}>Submit file</button>

            <h3>{JSON.stringify(message)}</h3>
        </form>
    )
}
