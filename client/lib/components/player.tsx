'use client'

import React from "react"
import usePlayer from "../hooks/usePlayer"
import dashjs from "dashjs";
import { MINIO_HOST } from "~/env";

export default function() {
    const [player] = React.useState(dashjs.MediaPlayer().create());
    const playerRef = React.useRef<HTMLAudioElement>(null);

    const { audio, visible } = usePlayer();

    React.useEffect(() => {
        if (player && audio?.manifest) {
            const url = MINIO_HOST + audio.manifest;
            player.initialize(playerRef.current, url, true, 0);
        }
    }, [player, audio]);

    return visible ? <audio controls ref={playerRef} /> : null
}
