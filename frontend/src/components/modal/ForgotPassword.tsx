"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React, {useState} from "react";
import Input from "@/components/Input";
import {useNotification} from "@/context/NotificationContext";
import {Button} from "@/components/Buttons";
import {useTranslations} from "next-intl";

export default function ForgotPassword() {
    const { activeModal, closeModal, } = useModal();
    const {addNotification} = useNotification();
    const [email, setEmail] = useState("");
    const t = useTranslations("auth.forgotPassword");
    const tSuccess = useTranslations("notifications.success");

    if (activeModal.type !== "forgot-password")
        return null;

    return (<ModalLayout onClose={closeModal} title={t("title")}>
        <Input id="email-forgot-password" value={email} onChange={setEmail} type="email" placeholder={t("email")}></Input>
        <Button className="w-full" onClick={() => {
            closeModal();
            setEmail("");
            addNotification(tSuccess("emailResetPassword"), "info");
        }}>{t("submit")}</Button>
    </ModalLayout>);
}
