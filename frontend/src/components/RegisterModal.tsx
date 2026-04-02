"use client";

import { useModal } from "@/context/ModalContext";
import ModalLayout from "@/components/LayoutModal";

export default function RegisterModal() {
    const {
        activeModal,
        closeModal,
    } = useModal();

    if (activeModal !== "register")
        return null;

    return (
        <ModalLayout onClose={closeModal}>
            <h2>Create Account</h2>

            <input
                type="email"
                placeholder="Email"
            />

            <input
                type="password"
                placeholder="Password"
            />

            <button>
                Register
            </button>
        </ModalLayout>
    );
}
