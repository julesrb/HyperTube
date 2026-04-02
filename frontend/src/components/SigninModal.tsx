"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";

export default function SigninModal() {
    const {
        activeModal,
        closeModal,
    } = useModal();

    if (activeModal !== "signin")
        return null;

    return (
        <ModalLayout onClose={closeModal}>
            <h2>Sign In</h2>

            <input
                type="email"
                placeholder="Email"
                className="field"
            />

            <input
                type="password"
                placeholder="Password"
                className="field"
            />

            <button className="button -action">
                Login
            </button>
        </ModalLayout>
    );
}