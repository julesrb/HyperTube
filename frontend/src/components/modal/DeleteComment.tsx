"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React from "react";
import {Button, SecondaryButton} from "@/components/Buttons";
import {useTranslations} from "next-intl";

export function DeleteCommentModal() {
    const {activeModal, closeModal,} = useModal();
    const t = useTranslations("modal.deleteComment");

    if (activeModal.type !== "delete-comment" || activeModal.commentId === undefined || activeModal.deleteComment === undefined)
        return null;

    return (<ModalLayout onClose={closeModal} title={t("title")}>
        <div className="flex gap-2 w-full">
            <Button
                className="w-full"
                onClick={() => {
                    if (activeModal.deleteComment && activeModal.commentId)
                        activeModal.deleteComment(activeModal.commentId);
                    closeModal();
                }}>{t("confirm")}</Button>
            <SecondaryButton
                className="w-full"
                onClick={closeModal}>{t("cancel")}</SecondaryButton>
        </div>
    </ModalLayout>);
}
