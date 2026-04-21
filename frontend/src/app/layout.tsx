import Navbar from "@/components/nav/Navbar";
import {ModalProvider} from "@/context/ModalContext";
import SigninModal from "@/components/modal/Signin";
import RegisterModal from "@/components/modal/Register";
import {GenreModal, FilterGenreModal} from "@/components/modal/Genre";
import "./fonts.css";
import "./globals.css";
import {NotificationProvider} from "@/context/NotificationContext";
import {NotificationList} from "@/components/NotificationList";
import React from "react";
import {AuthProvider} from "@/context/AuthContext";


export default function RootLayout({children,}: { children: React.ReactNode; }) {
    return (<html>
    <body>

    <AuthProvider>
        <NotificationProvider>

            <NotificationList/>

            <Navbar/>

            <SigninModal/>
            <RegisterModal/>
            <GenreModal/>
            <FilterGenreModal/>

            {children}

        </NotificationProvider>
    </AuthProvider>

    </body>
    </html>);
}
