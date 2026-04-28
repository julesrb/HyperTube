"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/modal/Layout";
import React from "react";
import {Button, SecondaryButton} from "@/components/Buttons";

export function DeleteCommentModal() {
    const {activeModal, closeModal,} = useModal();

    if (activeModal.type !== "delete-comment" || activeModal.commentId === undefined || activeModal.deleteComment === undefined)
        return null;

    return (<ModalLayout onClose={closeModal} title="Delete comment ?">
        <div className="flex gap-2 w-full">
            <Button
                className="w-full"
                onClick={() => {
                    if (activeModal.deleteComment && activeModal.commentId)
                        activeModal.deleteComment(activeModal.commentId);
                    closeModal();
                }}>OUI</Button>
            <SecondaryButton
                className="w-full"
                onClick={closeModal}>NON</SecondaryButton>
        </div>
    </ModalLayout>);
}
