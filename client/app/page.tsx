import styles from './page.module.css'

export default function Home() {
    console.log(process.env.GRPC_GATEWAY)

    return (
        <main className={styles.main}>
            <h1>
                {new Date().toString()}
            </h1>
        </main>
    )
}
