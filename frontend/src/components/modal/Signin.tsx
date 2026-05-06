"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useState} from "react";
import Input from "@/components/Input";
import {useAuth} from "@/context/AuthContext";
import {tUser, users} from "@/types/user";
import {tNotificationType, useNotification} from "@/context/NotificationContext";
import {Button, SmallButton} from "@/components/Buttons";
import {useTranslations} from "next-intl";

export default function Signin() {
    const {openModal, activeModal, closeModal,} = useModal();
    const {login} = useAuth();
    const {addNotification} = useNotification();
    const [password, setPassword] = useState("");
    const [username, setUsername] = useState("");
    const t = useTranslations("auth.signin");
    const tError = useTranslations("notifications.error");
    const tSuccess = useTranslations("notifications.success");

    if (activeModal.type !== "signin")
        return null;

    return (
        <ModalLayout onClose={closeModal} title={t("title")}>
            <Input id="username-signin" value={username} onChange={setUsername} type="username" placeholder={t("username")}></Input>
            <Input id="password-signin" value={password} onChange={setPassword} type="password" placeholder={t("password")}></Input>
            <div className="relative mb-4">
                <SmallButton className="absolute bottom-1" onClick={() => {
                    closeModal();
                    openModal({type: "forgot-password"});
                }}>{t("forgotPassword")}</SmallButton>
            </div>
            <Button className="h-8" onClick={() => handleLogin(login, addNotification, username, password, closeModal, tError("passwordIncorrect"), tSuccess("login"))}>{t("submit")}</Button>
            <div className="flex gap-2 mt-5">
                <span className="text-sm">{t("noAccount")}</span>
                <SmallButton onClick={() => {
                    closeModal();
                    openModal({type: "register"});
                }}>{t("register")}</SmallButton>
            </div>
        </ModalLayout>
    );
}

function handleLogin(login: (user: tUser, token: string) => void, addNotification: (message: string, type?: tNotificationType) => void, username: string, password: string, closeModal: () => void, passwordIncorrectMessage: string, loginSuccessMessage: string) {
    // const res = await fetch("/api/login", {
    //     method: "POST",
    //     body: JSON.stringify({
    //         email,
    //         password,
    //     }),
    // });
    //
    // const data = await res.json();
    // login(data.user, data.token);
    const findUser = users.filter(u => u.username === username);
    console.log("findUser", findUser);
    if (findUser.length > 0) {
        login(findUser[0], "coucou");
        closeModal();
        addNotification(loginSuccessMessage, "success");
    }
    else
        addNotification(passwordIncorrectMessage, "error");
}
