"use client";

import React from "react";
import {useTranslations} from "next-intl";
import useModal from "@/contexts/ModalContext";
import ModalLayout from "@/components/layout/ModalLayout";
import Form from "@/components/ui/Form";
import {patchApp, postNewApp} from "@/services/auth.service";
import {iApplication, tResponse} from "@/types/api";
import useNotification from "@/contexts/NotificationContext";
import {useQueryClient} from "@tanstack/react-query";
import {addQuery, updateQuery} from "@/hooks/useApiQuery";

export default function CreateApplicationModal() {
    const {activeModal, closeModal} = useModal();
    const tCreate = useTranslations("profile.application.createModal");
    const tEdit = useTranslations("profile.application.editModal");
    const {addNotification} = useNotification();
    const tSuccess = useTranslations("notifications.success");
    const queryClient = useQueryClient();

    if (activeModal.type !== "application")
        return null;

    const isCreate = activeModal.appId === undefined;
    const t = isCreate ? tCreate : tEdit;

    const handleCreateNewApp = (data: tResponse<iApplication>) => {
        if (data) {
            if (isCreate)
                addQuery(queryClient, ["applications", 0], data.data);
            else
                updateQuery(queryClient, ["applications"], data.data);
            closeModal();
            addNotification(tSuccess(isCreate ? "applicationCreated" : "applicationEdited"), "success");
        }
    }

    return (<ModalLayout onCloseAction={closeModal} title={t("title")}>
        <Form formType="application" request={isCreate ? postNewApp : patchApp} handleRequest={handleCreateNewApp} t={t}
              fields={["name", "redirect_uri"]} extraParam={activeModal.appId}/>
    </ModalLayout>);
}
