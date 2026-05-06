"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useState} from "react";
import Input from "@/components/Input";
import {tUser} from "@/types/user";
import {useAuth} from "@/context/AuthContext";
import {Button, SmallButton} from "@/components/Buttons";
import {useTranslations} from "next-intl";

export default function Register() {
    const {openModal, activeModal, closeModal} = useModal();
    const {login} = useAuth();
    const [email, setEmail] = useState("");
    const [firstname, setFirstname] = useState("");
    const [lastname, setLastname] = useState("");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const t = useTranslations("auth.register");

    if (activeModal.type !== "register")
        return null;

    return (
        <ModalLayout onClose={closeModal} title={t("title")}>
            <Input id="email-register" value={email} onChange={setEmail} type="email" placeholder={t("email")}></Input>

            <div className="flex gap-2">
                <Input id="firstname-register" value={firstname} onChange={setFirstname} type="firstname" placeholder={t("firstname")}></Input>
                <Input id="lastname-register" value={lastname} onChange={setLastname} type="lastname" placeholder={t("lastname")}></Input>
            </div>

            <Input id="username-register" value={username} onChange={setUsername} type="username" placeholder={t("username")} className={"max-w-2/3"}></Input>
            <Input id="password-register" value={password} onChange={setPassword} type="password" placeholder={t("password")} className={"max-w-2/3"}></Input>

            <Button className="h-8 mt-2" onClick={() =>
                handleRegister(login, username, email, firstname, lastname, password, closeModal)
            }>{t("submit")}</Button>

            <div className="flex gap-2 mt-5">
                <span className="text-sm">{t("haveAccount")}</span>
                <SmallButton onClick={() => {
                    closeModal();
                    openModal({type: "signin"});
                }}>{t("signIn")}</SmallButton>
            </div>
        </ModalLayout>
    );
}

function handleRegister(login: (user: tUser, token: string) => void, username: string, email: string, firstname: string, lastname: string, password: string, closeModal: () => void) {
    // const res = await fetch("/api/register", {
    //     method: "POST",
    //     body: JSON.stringify({
    //         email,
    //         username,
    //         firstname,
    //         lastname,
    //         password,
    //     }),
    // });
    //
    // const data = await res.json();
    // login(data.user, data.token);
    const user: tUser = {
        id: Date.now(),
        username: username,
        firstname: firstname,
        lastname: lastname,
        email: email,
        color: "purple",
        profile_picture: null,
        watch_history: [],
        joined_at: Date.now(),
    };
    login(user, "coucou");
    closeModal();
}
