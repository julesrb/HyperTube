import Navbar from "@/components/nav/Navbar";
import {ModalProvider} from "@/context/ModalContext";
import SigninModal from "@/components/modal/Signin";
import RegisterModal from "@/components/modal/Register";
import {GenreModal, FilterGenreModal} from "@/components/modal/Genre";
import "./fonts.css";
import "./globals.css";
import {NotificationProvider} from "@/context/NotificationContext";
import {NotificationList} from "@/components/Notifications";
import React from "react";
import {AuthProvider} from "@/context/AuthContext";
import ForgotPassword from "@/components/modal/ForgotPassword";
import {DeleteCommentModal} from "@/components/modal/DeleteComment";
import {hasLocale, NextIntlClientProvider} from 'next-intl';
import {routing} from "@/i18n/request";
import {notFound} from "next/navigation";

export default async function RootLayout({children, params}: {children: React.ReactNode, params: Promise<{locale: string}>}) {
    const {locale} = await params;
    if (!hasLocale(routing.locales, locale)) {
        notFound();
    }

    return (<html>
    <body>

    <NextIntlClientProvider>
        <AuthProvider>
            <NotificationProvider>
                <ModalProvider>
                    <NotificationList/>

                    <Navbar/>

                    <SigninModal/>
                    <RegisterModal/>
                    <GenreModal/>
                    <FilterGenreModal/>
                    <ForgotPassword/>
                    <DeleteCommentModal/>

                    {children}
                </ModalProvider>
            </NotificationProvider>
        </AuthProvider>
    </NextIntlClientProvider>

    </body>
    </html>);
}
