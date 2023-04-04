import styles from './page.module.css'

export default function Home() {

    return (
        <main className={styles.main}>
            <h1>
                {new Date().toString()}
            </h1>
        </main>
    )
}
