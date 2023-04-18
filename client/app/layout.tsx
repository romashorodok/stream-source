import { grpc } from '@improbable-eng/grpc-web'
import './globals.css'
import { NodeHttpTransport } from '@improbable-eng/grpc-web-node-http-transport'
import { PlayerProvider } from '~/lib/contexts/player-context'
import usePlayer from '~/lib/hooks/usePlayer'

export const metadata = {
    title: 'Create Next App',
    description: 'Generated by create next app',
}

grpc.setDefaultTransport(NodeHttpTransport())

export default function RootLayout({
    children,
}: {
    children: React.ReactNode
}) {
    return (
        <html lang="en">
            <body>
                <PlayerProvider>
                    {children}
                </PlayerProvider>
            </body>
        </html>
    )
}
