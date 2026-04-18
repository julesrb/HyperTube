import Navbar from "@/components/Navbar";
import {ModalProvider} from "@/context/ModalContext";
import SigninModal from "@/components/SigninModal";
import RegisterModal from "@/components/RegisterModal";
import "./fonts.css";
import "./globals.css";
import {NotificationProvider} from "@/context/NotificationContext";
import {NotificationList} from "@/components/NotificationList";


export default function RootLayout({children,}: { children: React.ReactNode; }) {
    return (<html>
    <body>

    <ModalProvider>
        <NotificationProvider>

            <NotificationList/>

            <Navbar/>

            <SigninModal/>
            <RegisterModal/>

            {children}

        </NotificationProvider>
    </ModalProvider>

    </body>
    </html>);
}
