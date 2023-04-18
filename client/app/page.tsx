import { AudioService } from '~/lib/services/audio.service';
import styles from './page.module.css'
import AudioCard from '~/lib/components/audio-card';
import Player from '~/lib/components/player';

const audioService = new AudioService();

export default async function Home() {

    const { audiosList } = await audioService.listAudios()

    return (
        <main className={styles.main}>
            {audiosList.map(({ audio, manifest }) => (
                <AudioCard audio={audio} manifest={manifest} />
            ))}

            <Player/>

            <h1>
                {new Date().toString()}
            </h1>
        </main>
    )
}
