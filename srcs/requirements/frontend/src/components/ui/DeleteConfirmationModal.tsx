"use client";

import React from "react";
import {useTranslations} from "next-intl";
import useModal from "@/contexts/ModalContext";
import ModalLayout from "@/components/layout/ModalLayout";
import Button from "@/components/ui/Button/Button";
import SecondaryButton from "@/components/ui/Button/SecondaryButton";
import useApiMutation from "@/hooks/useApiMutation";

export default function DeleteConfirmationModal() {
    const {activeModal, closeModal} = useModal();
    const t = useTranslations("modal.deleteConfirmation");
    const {execute} = useApiMutation();

    if (activeModal.type !== "delete-confirmation" || activeModal.deleteObjId === undefined || activeModal.deleteFunc === undefined)
        return null;

    return (<ModalLayout onCloseAction={closeModal} title={t("title")}>
        <div className="flex gap-2 w-full">
            <Button className="w-full" onClick={async () => {
                if (activeModal.deleteFunc && activeModal.deleteObjId)
                    await execute(() => activeModal.deleteFunc!(activeModal.deleteObjId ?? 0));
                closeModal();
            }}>{t("confirm")}</Button>
            <SecondaryButton className="w-full" onClick={closeModal}>{t("cancel")}</SecondaryButton>
        </div>
    </ModalLayout>);
}
