import React, {useEffect, useRef, useState} from "react";
import {useTranslations} from "next-intl";
import {tOauthService} from "@/types/user";
import {OAuthIcon} from "@/components/Icons";
import {usePathname} from "@/i18n/navigation";
import useApiMutation from "@/hooks/useApiMutation";
import {tResponse} from "@/types/api";
import useAuth from "@/contexts/AuthContext";
import Button from "@/components/ui/Button/Button";
import {handleOauth} from "@/services/auth.service";
import TextButton from "@/components/ui/Button/TextButton";
import Input from "@/components/ui/Input";
import useResponsiveSize from "@/hooks/useResponsiveSize";

export type fieldType = "email" | "login" | "first_name" |  "last_name" | "username" | "password" | "current-password" | "new-password" | "confirm-new-password" | "name" | "redirect_uri";
type formType = "auth" | "update" |  "signin" | "register" | "send-email-reset-password" | "set-new-password" | "application";

export default function Form<T>({formType, request, handleRequest, t, fields, handleForgotPassword, extraParam}: {formType: formType, request: (locale: string, data: string[], extraParam?: string | number) => Promise<tResponse<T>>, handleRequest: (data: tResponse<T>) => void, t: (key: string) => string, fields: fieldType[], handleForgotPassword?: () => void, extraParam?: string | number}) {
    const [fieldsValue, setFieldsValue] = useState<string[]>(Array(fields.length).fill(""));
    const [errors, setErrors] = useState<Record<string, string>>({});
    const [disableBtn, setDisableBtn] = useState(false);
    const tError = useTranslations("validationErrors");
    const [focusedIndex, setFocusedIndex] = useState((formType === "update" || formType === "auth")? -1 : 0);
    const {execute} = useApiMutation(setErrors, setFocusedIndex, formType, fields);
    const fieldRefs = useRef<HTMLInputElement[]>([]);
    const showOAuth = formType === "signin" || formType === "register";

    useEffect(() => {
        if (focusedIndex >= 0)
            fieldRefs.current[focusedIndex]?.focus();
    }, [focusedIndex]);

    const getId = (field: fieldType | number) => (typeof field  === "number" ? fields[field] : field) + "-" + formType;
    const hasError = (error: Record<string, string>) => Object.keys(error).length > 0 && Object.values(error).some((v) => !!v);
    const newSetterErrorUtils = (field: fieldType | number, error?: string) => newSetterError({[getId(field)]: error ? tError(error) : ""});

    const newSetterError = (value: Record<string, string>) => {
        setErrors((prevErrors) => {
            const newErrors = {...prevErrors, ...value};
            const field = formType === "application" ? "name" : "email";
            if (prevErrors[getId(field)] === tError("atLeastOneField") && !value[getId(field)])
                newErrors[getId(field)] = "";
            setDisableBtn(hasError(newErrors));
            return newErrors;
        });
    };

    useEffect(() => {
        if (formType === "signin" && errors[getId("login")] == tError("invalidLoginOrPassword"))
            newSetterErrorUtils("login");

        if (formType === "auth" && fieldsValue[0]) {
            if (fieldsValue[0] == fieldsValue[1])
                newSetterErrorUtils("new-password", "passwordSameAsOld");
            else if (errors[getId("new-password")] == tError("passwordSameAsOld"))
                newSetterErrorUtils("new-password")
        }

        if ((formType === "auth" || formType === "set-new-password") && fieldsValue.at(-1) && fieldsValue.at(-2)) {
            if (fieldsValue.at(-1) != fieldsValue.at(-2))
                newSetterErrorUtils("confirm-new-password", "passwordsDontMatch");
            else if (errors[getId("confirm-new-password")] == tError("passwordsDontMatch"))
                newSetterErrorUtils("confirm-new-password");
        }
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [fieldsValue]);

    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>, index: number) => {
        if (e.key !== "Enter")
            return ;
        e.preventDefault();
        const isLastField = index === fields.length - 1;
        if (isLastField) {
            if (hasError(errors)) {
                for (let i = 0; i < fields.length; i++) {
                    if (errors[getId(i)]) {
                        setFocusedIndex(i);
                        return ;
                    }
                }
            }
            onSubmit();
        }
        else if (index + 1 === focusedIndex)
            fieldRefs.current[focusedIndex]?.focus();
        else
            setFocusedIndex(index + 1);
    }

    const onSubmit = () => {
        const makeRequest = async () => {
            return await execute((locale) => request(locale, fieldsValue, extraParam));
        };

        if (disableBtn)
            return ;
        const requiredErrors: Record<string, string> = {};
        const isAllFieldRequired = !(formType === "update" || (formType === "application" && extraParam != undefined));

        if (!isAllFieldRequired && fieldsValue.filter((v) => v.trim().length !== 0).length === 0) {
            newSetterErrorUtils(0, "atLeastOneField")
            setFocusedIndex(0);
        } else if (isAllFieldRequired && fieldsValue.filter((v) => v.trim().length === 0).length !== 0) {
            let focusIsSet = false;

            fieldsValue.map((val: string, idx: number) => {
                if (val.trim().length === 0) {
                    requiredErrors[getId(idx)] = tError("requiredField");
                    if (!focusIsSet) {
                        focusIsSet = true;
                        setFocusedIndex(idx);
                    }
                }
            })
            newSetterError(requiredErrors);
        } else {
            makeRequest().then((data) => {
                if (data) {
                    handleRequest(data);
                    setFieldsValue(Array(fields.length).fill(""));
                }
            })
            fieldRefs.current[focusedIndex]?.blur();
            setFocusedIndex(-1);
        }
    };

    const RenderInput = (type: fieldType, idx: number, className?: string) =>
        <Input key={idx} id={getId(type)} type={type.includes("password") ? "password" : "text"} placeholder={t(type)}
               value={fieldsValue[idx]} idx={idx} onChange={setFieldsValue} className={className + ((type === "username" || (formType === "register" && type === "password")) ? " max-w-2/3" : "")}
               requestErrorMessage={errors[getId(type)]} setErrorsMessage={newSetterError} ref={(el: HTMLInputElement) => {fieldRefs.current[idx] = el;}}
               onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => handleKeyDown(e, idx)} />;

    return (<form id={formType} className="w-full" onSubmit={(e) => {e.preventDefault(); onSubmit();}}>
        {fields.map((field, idx) => {
            if (field === "first_name") {
                return (<div key="div" className="flex gap-2">
                    {RenderInput(field, idx)}
                    {RenderInput(fields[idx + 1], idx + 1)}
                </div>);
            } else if (field !== "last_name")
                return RenderInput(field, idx);
        })}
        {handleForgotPassword && <div className={"relative mb-4" + (errors["password-signin"] ? " pt-3" : "")}>
            <TextButton className="absolute bottom-1" onClick={handleForgotPassword}>{t("forgotPassword")}</TextButton>
        </div>}
        <div className="flex gap-2">
            <Button onClick={onSubmit} disabled={disableBtn} className={formType.includes("password") ? "w-full" : ""}>{t("submit")}</Button>
            {showOAuth && <OauthServices oauth="42" title={t("oAuth")} />}
            {showOAuth && <OauthServices oauth="github" title={t("oAuth")} />}
            {showOAuth && <OauthServices oauth="gitlab" title={t("oAuth")} />}
        </div>
    </form>)
}

function OauthServices({oauth, title}: {oauth: tOauthService, title: string}) {
    const {callbackUrl} = useAuth();
    const pathname = usePathname();
    const size = useResponsiveSize();
    const iconSize = size === "xs" ? 23 : 30;

    return (<button type="button" title={title + " " + oauth} onClick={() => handleOauth(oauth, callbackUrl || pathname)} className="flex items-center justify-center size-10 hover:bg-black-hover bg-black">
        <OAuthIcon oauth={oauth} size={iconSize}/>
    </button>)
}
