"use client";

import React from "react";
import {useTranslations} from "next-intl";
import useModal from "@/contexts/ModalContext";
import useAuth from "@/contexts/AuthContext";
import useNotification from "@/contexts/NotificationContext";
import ModalLayout from "@/components/layout/ModalLayout";
import Form from "@/components/ui/Form";
import {postSetNewPassword} from "@/services/auth.service";

export default function ResetPasswordModal() {
    const {setCallbackUrl} = useAuth();
    const {activeModal, openModal, closeModal} = useModal();
    const {addNotification} = useNotification();
    const t = useTranslations("auth.resetPassword");
    const tSuccess = useTranslations("notifications.success");

    if (activeModal.type !== "set-new-password" || activeModal.token === undefined)
        return null;

    const handleNewPassword = () => {
        closeModal();
        addNotification(tSuccess("newPasswordSet"), "success");
        openModal({type: "signin"});
    }

    return (<ModalLayout onCloseAction={() => {closeModal(); setCallbackUrl(undefined);}} title={t("title")}>
        <Form formType="set-new-password" request={postSetNewPassword} handleRequest={handleNewPassword} fields={["new-password", "confirm-new-password"]} t={t} extraParam={activeModal.token} />
    </ModalLayout>);
}
