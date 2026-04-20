import Navbar from "@/components/nav/Navbar";
import {ModalProvider} from "@/context/ModalContext";
import SigninModal from "@/components/modal/Signin";
import RegisterModal from "@/components/modal/Register";
import GenreModal from "@/components/modal/Genre";
import "./fonts.css";
import "./globals.css";
import {NotificationProvider} from "@/context/NotificationContext";
import {NotificationList} from "@/components/NotificationList";
import React from "react";


export default function RootLayout({children,}: { children: React.ReactNode; }) {
    return (<html>
    <body>

    <ModalProvider>
        <NotificationProvider>

            <NotificationList/>

            <Navbar/>

            <SigninModal/>
            <RegisterModal/>
            <GenreModal/>

            {children}

        </NotificationProvider>
    </ModalProvider>

    </body>
    </html>);
}
