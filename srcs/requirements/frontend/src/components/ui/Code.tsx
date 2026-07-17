import IconButton from "@/components/ui/Button/IconButton";
import {CopyIcon} from "@/components/Icons";
import useNotification from "@/contexts/NotificationContext";
import {useTranslations} from "next-intl";

export default function Code({children, label}: {children: string, label: string}) {
    const {addNotification} = useNotification();
    const tSuccess = useTranslations("notifications.success");

    const copyText = () => {
        navigator.clipboard.writeText(children).then(() => {
            addNotification(tSuccess("textCopied", {label: label}), "success");
        })
    }

    return (<div onClick={copyText} className={"flex truncate justify-between items-center gap-2 font-hairline bg-white-loading px-2 py-1 overflow-x-hidden" + (children ? " hover:cursor-pointer" : "")}>
        <p className="truncate">{label}: {children}</p>
        <IconButton>{(color) => <CopyIcon color={color}/>}</IconButton>
    </div>);
}
