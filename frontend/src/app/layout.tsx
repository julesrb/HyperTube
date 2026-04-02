import Navbar from "@/components/Navbar";
import {ModalProvider} from "@/context/ModalContext";
import SigninModal from "@/components/SigninModal";
import RegisterModal from "@/components/RegisterModal";
import "./fonts.css";
import "./globals.css";


export default function RootLayout({children,}: { children: React.ReactNode; }) {
    return (<html>
        <body>

        <ModalProvider>

            <Navbar/>

            <SigninModal />
            <RegisterModal />

            {children}

        </ModalProvider>

        </body>
        </html>);
}
