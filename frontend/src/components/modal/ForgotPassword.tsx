"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useState} from "react";
import Input from "@/components/Input";
import {Button} from "@/components/Button";
import {useNotification} from "@/context/NotificationContext";
import {successMessages} from "@/types/message";

export default function ForgotPassword() {
    const { activeModal, closeModal, } = useModal();
    const {addNotification} = useNotification();
    const [email, setEmail] = useState("");

    if (activeModal.type !== "forgot-password")
        return null;

    return (<ModalLayout onClose={closeModal} title="Réinitialiser le mot de passe">
        <Input id="email-forgot-password" value={email} onChange={setEmail} type="email" placeholder="Email"></Input>
        <Button className="w-full" onClick={() => {
            closeModal();
            setEmail("");
            addNotification(successMessages.emailResetPassword, "info");
        }}>Envoyer</Button>
    </ModalLayout>);
}
