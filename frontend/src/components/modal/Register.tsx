"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useState} from "react";
import Input from "@/components/Input";
import {tUser} from "@/types/user";
import {useAuth} from "@/context/AuthContext";
import {Button} from "@/components/Buttons";

export default function Register() {
    const {activeModal, closeModal} = useModal();
    const {login} = useAuth();
    const [email, setEmail] = useState("");
    const [firstname, setFirstname] = useState("");
    const [lastname, setLastname] = useState("");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");

    if (activeModal.type !== "register")
        return null;

    return (
        <ModalLayout onClose={closeModal} title="Join Hypertube">
            <Input id="email-register" value={email} onChange={setEmail} type="email" placeholder="Email"></Input>

            <div className="flex gap-2">
                <Input id="firstname-register" value={firstname} onChange={setFirstname} type="firstname" placeholder="Firstname"></Input>
                <Input id="lastname-register" value={lastname} onChange={setLastname} type="lastname" placeholder="Lastname"></Input>
            </div>

            <Input id="username-register" value={username} onChange={setUsername} type="username" placeholder="Username" className={"max-w-2/3"}></Input>
            <Input id="password-register" value={password} onChange={setPassword} type="password" placeholder="Password" className={"max-w-2/3"}></Input>

            <Button className="h-8 mt-2" onClick={() =>
                handleRegister(login, username, email, firstname, lastname, password, closeModal)
            }>Sign Up</Button>
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
        film_history: [],
        joined_at: Date.now(),
    };
    login(user, "coucou");
    closeModal();
}
