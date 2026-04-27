"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useState} from "react";
import Input from "@/components/Input";
import {useAuth} from "@/context/AuthContext";
import {tUser, users} from "@/types/user";
import {tNotificationType, useNotification} from "@/context/NotificationContext";
import {errorMessages, successMessages} from "@/types/message";
import {Button, SmallButton} from "@/components/Buttons";

// export default function SigninModal() {
//     const {activeModal, closeModal,} = useModal();
//
//     if (activeModal.type !== "signin")
//         return null;
//
//     return (
//         <ModalLayout onClose={closeModal} defaultLayout={false}>
//             <div className="absolute top-0 right-1/10 flex gap-2 p-2 bg-white custom-shadow-s border" onClick={(e) => e.stopPropagation()}>
//                 <button onClick={closeModal}><CrossIcon /></button>
//                 <Input type="email" placeholder="Email"></Input>
//                 <Input type="password" placeholder="Password"></Input>
//
//                 <Button className="h-8">Sign In</Button>
//             </div>
//         </ModalLayout>
//     );
// }

export default function Signin() {
    const {openModal, activeModal, closeModal,} = useModal();
    const {login} = useAuth();
    const {addNotification} = useNotification();
    const [password, setPassword] = useState("");
    const [username, setUsername] = useState("");

    if (activeModal.type !== "signin")
        return null;

    return (
        <ModalLayout onClose={closeModal} title="Signin">
            <Input id="username-signin" value={username} onChange={setUsername} type="username" placeholder="Username"></Input>
            <Input id="password-signin" value={password} onChange={setPassword} type="password" placeholder="Password"></Input>
            <div className="relative mb-4">
                <SmallButton className="absolute bottom-1" onClick={() => {
                    closeModal();
                    openModal({type: "forgot-password"});
                }}>Forgotten?</SmallButton>
            </div>
            <Button className="h-8" onClick={() => handleLogin(login, addNotification, username, password, closeModal)}>Sign In</Button>
            <div className="flex gap-2 mt-5">
                <span className="text-sm">Pas encore de compte?</span>
                <SmallButton onClick={() => {
                    closeModal();
                    openModal({type: "register"});
                }}>Inscrivez-vous</SmallButton>
            </div>
        </ModalLayout>
    );
}

function handleLogin(login: (user: tUser, token: string) => void, addNotification: (message: string, type?: tNotificationType) => void, username: string, password: string, closeModal: () => void) {
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
        addNotification(successMessages.login, "success");
    }
    else
        addNotification(errorMessages.passwordIncorrect, "error");
}
