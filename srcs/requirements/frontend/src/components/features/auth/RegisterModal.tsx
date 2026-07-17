"use client";

import React from "react";
import {useTranslations} from "next-intl";
import useModal from "@/contexts/ModalContext";
import AuthModalLayout from "@/components/layout/AuthModalLayout";

export default function RegisterModal() {
    const {activeModal} = useModal();
    const t = useTranslations("auth.register");

    if (activeModal.type !== "register")
        return null;

    return (<AuthModalLayout type="register" t={t} activeModal={activeModal}/>)
}
