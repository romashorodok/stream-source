'use client'

import { Audio } from "pb/ts/audio/v1/audio_service_pb"
import usePlayer from "../hooks/usePlayer";

type Props = {
    audio: Audio.AsObject;
    manifest: string;
}

export default function({ audio, manifest }: Props) {
    const { setAudio } = usePlayer();

    return (
        <div>
            <h4>{audio.title}</h4>
            <button onClick={() => setAudio({ ...audio, manifest })}>Play</button>
        </div>
    )
}
