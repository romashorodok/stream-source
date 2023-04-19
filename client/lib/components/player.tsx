'use client'

import React from "react"
import usePlayer from "../hooks/usePlayer"
import dashjs from "dashjs";
import { MINIO_HOST } from "~/env";
import * as Slider from '@radix-ui/react-slider';
import { PlayButton } from "./audio-card";
import { SpeakerLoudIcon } from "@radix-ui/react-icons";


function formatAudioDuration(duration: number) {
    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration - (hours * 3600)) / 60);
    const seconds = Math.floor(duration - (hours * 3600) - (minutes * 60));

    let formattedDuration = "";

    if (hours > 0) {
        formattedDuration += hours + ":";
    }

    if (minutes > 0) {
        formattedDuration += minutes + ":";
    } else {
        formattedDuration += "00:";
    }

    formattedDuration += seconds;

    return formattedDuration;
}

export default function({ className }: { className?: string }) {
    const [player] = React.useState(dashjs.MediaPlayer().create());

    const [time, setTime] = React.useState<number>();
    const [duration, setDuration] = React.useState<number>();
    const [volume, setVolume] = React.useState<number>();

    const playerRef = React.useRef<HTMLAudioElement>(null);

    const { setPlayer, audio, setPlaying, playing } = usePlayer();

    const visible = React.useMemo(() =>
        audio ?
            Object.values(audio)
                .every((value) => value !== undefined && value !== null)
            : false
        , [audio]);

    React.useEffect(() => {
        player.on('playbackPaused', stopPlayingState);
        player.on('playbackPlaying', startPlayingState);

        playerRef?.current?.addEventListener('timeupdate', onUpdate);
        playerRef?.current?.addEventListener('volumechange', onVolumeChange);

        if (audio?.manifest) {
            player?.initialize(playerRef.current, MINIO_HOST + audio.manifest, true, 0);
            setPlayer(player);
        }

        return () => {
            player.off("playbackPaused", stopPlayingState);
            player.off("playbackPlaying", startPlayingState);

            playerRef?.current?.removeEventListener('timeupdate', onUpdate);
            playerRef?.current?.removeEventListener('volumechange', onVolumeChange);
        }
    }, [audio]);

    function onVolumeChange(event: any) {
        setVolume(event.target.volume);
    }

    function onUpdate(event: any) {
        setTime(event.target.currentTime);
    }

    function startPlayingState() {
        setPlaying(true);
        setDuration(player.duration());
        setVolume(player.getVolume());
    }

    function stopPlayingState() {
        setPlaying(false);
    }

    return visible
        ? (
            <div className="flex flex-row gap-2 pt-2">
                <audio ref={playerRef} />

                <div>
                    <div className="flex justify-center">
                        <PlayButton
                            className="flex items-center justify-center bg-white bottom-[6px] right-[6px] rounded-3xl bg-black w-[33px] h-[33px]"
                            stopIconClassName="fill-curren text-black tw-[24px] h-[24px]"
                            playIconClassName="fill-curren text-black tw-[28px] h-[28px]"
                            active={playing}
                            onPlayClick={() => player.play()}
                            onPauseClick={() => player.pause()}
                        />

                    </div>
                    <div className="flex items-center space-x-4">
                        <p>{formatAudioDuration(time)}</p>
                        <Slider.Root
                            className={`relative flex items-center select-none touch-none w-[200px] h-5 ` + className}
                            value={[time]}
                            max={duration}
                            onValueChange={([duration]) => playerRef.current.currentTime = duration}
                            defaultValue={[0]}
                            step={1}
                            aria-label="Volume" >
                            <Slider.Track className="bg-blackA10 relative grow rounded-full h-[3px]">
                                <Slider.Range className="absolute bg-white rounded-full h-full" />
                            </Slider.Track>
                            <Slider.Thumb className="block w-5 h-5 bg-white shadow-[0_2px_10px] shadow-blackA7 rounded-[10px] hover:bg-violet3 focus:outline-none focus:shadow-[0_0_0_5px] focus:shadow-blackA8" />

                        </Slider.Root>
                        <p>{formatAudioDuration(duration)}</p>
                    </div>
                </div>
                <div className="flex items-center space-x-4 ml-5">
                    <SpeakerLoudIcon />

                    <Slider.Root
                        className={`relative flex items-center select-none touch-none w-[80px] h-5 ` + className}
                        value={[volume]}
                        max={1}
                        onValueChange={([volume]) => player.setVolume(volume)}
                        defaultValue={[0]}
                        step={0.1}
                        aria-label="Volume" >
                        <Slider.Track className="bg-blackA10 relative grow rounded-full h-[3px]">
                            <Slider.Range className="absolute bg-white rounded-full h-full" />
                        </Slider.Track>
                        <Slider.Thumb className="block w-5 h-5 bg-white shadow-[0_2px_10px] shadow-blackA7 rounded-[10px] hover:bg-violet3 focus:outline-none focus:shadow-[0_0_0_5px] focus:shadow-blackA8" />
                    </Slider.Root>
                </div>
            </div>
        )
        : null
}

