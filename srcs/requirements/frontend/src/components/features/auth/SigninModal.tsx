"use client";

import React from "react";
import {useTranslations} from "next-intl";
import useModal from "@/contexts/ModalContext";
import AuthModalLayout from "@/components/layout/AuthModalLayout";

export default function SigninModal() {
    const {openModal, activeModal, closeModal} = useModal();
    const t = useTranslations("auth.signin");

    if (activeModal.type !== "signin")
        return null;

    const handleForgotPassword = () => {
        closeModal();
        openModal({type: "send-email-forgot-password"});
    };

    return (<AuthModalLayout type="signin" t={t} handleForgotPassword={handleForgotPassword} activeModal={activeModal}/>)
}
