'use client'

import React from "react"
import { PlayerContext } from "../contexts/player-context"

export default function usePlayer() {

    const context = React.useContext(PlayerContext);

    const visible = React.useMemo(() =>
        context.audio ?
            Object.values(context.audio)
                .every((value) => value !== undefined && value !== null)
            : false
        , [context.audio]);

    return { visible, setAudio: context.setAudio, audio: context.audio }
}
