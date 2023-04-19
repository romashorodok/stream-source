import { AudioService } from '~/lib/services/audio.service';
import styles from './page.module.scss'
import AudioCard from '~/lib/components/audio-card';
import Player from '~/lib/components/player';

const audioService = new AudioService();

export default async function Home() {

    const { audiosList } = await audioService.listAudios()

    return (
        <main className={`grid grid-cols-3 h-screen ${styles.grid_area}`}>
            <div className={`grid p-8 overflow-auto ${styles.grid_audio_cards} ${styles.area_content}`}>
                {audiosList.map(({ audio, manifest }) => (
                    <AudioCard audio={audio} manifest={manifest} />
                ))}
            </div>
            <div className={`flex justify-center bg-primary ${styles.area_player}`}>
                <Player className="w-[400px]" />
            </div>
        </main>
    )
}

