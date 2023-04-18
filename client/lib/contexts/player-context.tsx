'use client'

import React, { useState } from "react";
import { PlayableAudio } from "../types/Player";

type PlayableAudioContext = {
    audio: PlayableAudio;
    setAudio: React.Dispatch<React.SetStateAction<PlayableAudio>>
}

export const PlayerContext = React.createContext<PlayableAudioContext>(undefined);

export function PlayerProvider({ children }: React.PropsWithChildren) {
    const [audio, setAudio] = React.useState<PlayableAudio>();

    const context = React.useMemo(() => ({
        audio, setAudio
    }), [audio]);

    return (
        <PlayerContext.Provider value={context}>
            {children}
        </PlayerContext.Provider>
    )
}

