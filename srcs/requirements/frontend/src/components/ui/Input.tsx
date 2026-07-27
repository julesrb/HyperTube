import React, {useState} from "react";
import {EyeIcon} from "@/components/Icons";
import {useTranslations} from "next-intl";
import IconButton from "@/components/ui/Button/IconButton";

export default function Input(
    {id, type, placeholder, value, onChange, idx, className, requestErrorMessage, setErrorsMessage, ref, onKeyDown}:
    {id: string, type: string, placeholder: string, value: string, onChange: React.Dispatch<React.SetStateAction<string[]>>, idx: number, className?: string, requestErrorMessage?: string, setErrorsMessage?: (errorMsg: Record<string, string>) => void, ref?: (el: HTMLInputElement) => void, onKeyDown?: (e: React.KeyboardEvent<HTMLInputElement>) => void}
) {
    const isPassword = type === "password";
    const t = useTranslations("validationErrors");
    const [isPasswordVisible, setIsPasswordVisible] = useState(false);
    const usernameRegex = /^[a-zA-Z0-9_]+$/;
    const emailRegex = /^(?=.{1,64}@)(?!.*\.\.)([a-zA-Z0-9_+-]+(?:\.[a-zA-Z0-9_+-]+)*)@(?=.{1,253}$)(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$/;
    const nameRegex = /^\p{L}+(?:[ '-]\p{L}+)*$/u;
    const urlRegex = /^https?:\/\/[^\s]+$/;

    const handleTogglePasswordVisibility = () => setIsPasswordVisible(!isPasswordVisible);

    const handleFieldVerification = (e: React.ChangeEvent<HTMLInputElement, HTMLInputElement>) => {
        const newValue = e.target.value;

        if (setErrorsMessage) {
            let message = "";

            const tNewValue = newValue.trim();
            if (id.includes("email")) {
                if (tNewValue && !emailRegex.test(tNewValue))
                    message = t("emailInvalid");
            } else if (id.includes("last_name") || id.includes("first_name")) {
                const field = id.includes("last_name") ? "firstname" : "lastname";
                if (tNewValue.length > 30)
                    message = t(field + "TooLong");
                else if (tNewValue && !nameRegex.test(tNewValue))
                    message = t(field + "Invalid");
            } else if (id.includes("username")) {
                if (tNewValue && tNewValue.length < 3)
                    message = t("usernameTooShort");
                else if (tNewValue.length > 32)
                    message = t("usernameTooLong");
                else if (newValue && !usernameRegex.test(newValue))
                    message = t("usernameInvalid");
            } else if (id.includes("password")) {
                if (newValue && newValue.length < 8)
                    message = t("passwordTooShort");
                else if (newValue.length > 72)
                    message = t("passwordTooLong");
            } else if (id.includes("redirect_uri")) {
                if (tNewValue && !tNewValue.startsWith("http://") && !tNewValue.startsWith("https://"))
                    message = t("redirectURIStartHTTP");
                else if (tNewValue && !urlRegex.test(tNewValue))
                    message = t("redirectURIInvalid");
            }
            setErrorsMessage({[id]: message});
        }
        onChange((prev) => {
            const updated = [...prev];
            updated[idx] = newValue;
            return updated;
        });
    }

    return (<div className={"flex flex-col w-full h-16 relative " + className}>
        <input id={id} type={isPasswordVisible && isPassword ? "text" : type} placeholder=""
               value={value} onChange={handleFieldVerification} onKeyDown={onKeyDown} ref={ref}
               className={"peer py-4 m-0 w-full h-8 bg-white  border-b focus:border-b-2 " + (requestErrorMessage ? "border-b-red text-red" : "text-black")}
        />
        <label htmlFor={id}
               className={"pointer-events-none uppercase absolute text-xs font-light bottom-15\
                   peer-focus:text-xs peer-focus:font-body peer-focus:font-light peer-focus:bottom-15\
                   peer-placeholder-shown:font-condensed peer-placeholder-shown:tracking-wide peer-placeholder-shown:bottom-9 peer-placeholder-shown:text-2xl" + (requestErrorMessage ? " text-red" : "")}>{placeholder}</label>
        {isPassword && (<button type="button" className="absolute right-0 top-1" onClick={handleTogglePasswordVisibility}><EyeIcon crossed={isPasswordVisible} color={requestErrorMessage ? "red" : "black"} /></button>)}
        {isPassword && (<IconButton color={requestErrorMessage ? "red" : "black"} className="absolute right-0 top-1" onClick={handleTogglePasswordVisibility}>{(color: string) => <EyeIcon crossed={isPasswordVisible} color={color}/>}</IconButton>)}
        {requestErrorMessage && <span className="max-w-80 text-xs text-red overflow-x-scroll text-nowrap scrollbar-hide">{requestErrorMessage}</span>}
    </div>
    );
}
