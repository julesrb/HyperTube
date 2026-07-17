import {useTranslations} from "next-intl";
import {useCallback} from "react";
import Link from "next/link";
import {usePathname, useRouter} from "@/i18n/navigation";
import {ApiError} from "@/services/ApiError";
import useModal from "@/contexts/ModalContext";
import useNotification from "@/contexts/NotificationContext";
import useAuth from "@/contexts/AuthContext";
import SmallText from "@/components/ui/SmallText";

export default function useHandleError() {
    const {openModal} = useModal();
    const {addNotification} = useNotification();
    const tError = useTranslations("notifications.error");
    const {setCallbackUrl, logout} = useAuth();
    const pathname = usePathname();
    const router = useRouter();

    return useCallback((error: ApiError, translation: "Film" | "User") => {
        if (error instanceof ApiError) {
            if (error.status === 401) {
                setCallbackUrl(pathname);
                openModal({type: "signin"});
                logout();
                return (<button className="w-full hover:underline" onClick={() =>
                    openModal({type: "signin"})
                }><SmallText>{tError("loginRequired")}</SmallText></button>);
            } else if (error.status === 404)
                addNotification(tError("notFound" + translation), "error");
                router.push("/");
                return (<SmallText>{tError("notFound" + translation)}</SmallText>);
        } else
            addNotification(tError("network"), "error");
        return (<Link href="/"><p className="text-center italic text-red">{tError("unknown")}</p></Link>);
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pathname, tError]);
}
