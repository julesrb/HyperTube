"use client";

import React from "react";
import {useTranslations} from "next-intl";
import useModal from "@/contexts/ModalContext";
import useAuth from "@/contexts/AuthContext";
import useNotification from "@/contexts/NotificationContext";
import ModalLayout from "@/components/layout/ModalLayout";
import Form from "@/components/ui/Form";
import {postResetPassword} from "@/services/auth.service";

export default function ForgotPasswordModal() {
    const {setCallbackUrl} = useAuth();
    const {activeModal, closeModal} = useModal();
    const {addNotification} = useNotification();
    const t = useTranslations("auth.forgotPassword");
    const tSuccess = useTranslations("notifications.success");

    if (activeModal.type !== "send-email-forgot-password")
        return null;

    const handleResetPassword = () => {
        setCallbackUrl(undefined);
        closeModal();
        addNotification(tSuccess("emailResetPassword"), "warning");
    }

    return (<ModalLayout onCloseAction={() => {closeModal(); setCallbackUrl(undefined);}} title={t("title")}>
        <Form formType="send-email-reset-password" request={postResetPassword} handleRequest={handleResetPassword} fields={["email"]} t={t} />
    </ModalLayout>);
}
