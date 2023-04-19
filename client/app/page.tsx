import { AudioService } from '~/lib/services/audio.service';
import styles from './page.module.scss'
import AudioCard from '~/lib/components/audio-card';

const audioService = new AudioService();

export default async function Home() {

    const { audiosList } = await audioService.listAudios()

    return (
        <div className={`grid p-8 gap-8 overflow-auto ${styles.grid_audio_cards}`}>
            {audiosList.map(({ audio, manifest }, index) => (
                <AudioCard key={index} audio={audio} manifest={manifest} />
            ))}
        </div>
    )
}

