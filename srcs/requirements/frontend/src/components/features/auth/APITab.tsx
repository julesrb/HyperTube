import Pagination from "@/components/ui/Pagination";
import computeTotalPage from "@/utils/computeTotalPage";
import React, {useEffect, useState} from "react";
import {deleteApp, useApplications} from "@/services/auth.service";
import {iApplication} from "@/types/api";
import SmallText from "@/components/ui/SmallText";
import {useLocale, useTranslations} from "next-intl";
import Button from "@/components/ui/Button/Button";
import useModal, {ModalState} from "@/contexts/ModalContext";
import IconButton from "@/components/ui/Button/IconButton";
import {TrashIcon} from "@/components/Icons";
import Code from "@/components/ui/Code";
import Label from "@/components/ui/Label";
import {useQueryClient} from "@tanstack/react-query";
import {removeQuery} from "@/hooks/useApiQuery";
import TextButton from "@/components/ui/Button/TextButton";
import {useRouter} from "@/i18n/navigation";

export default function APITab() {
    const locale = useLocale();
    const [index, setIndex] = useState(0);
    const {data} = useApplications(index);
    const [apps, setApps] = useState<iApplication[]>([]);
    const totalPage = computeTotalPage(data);
    const {openModal} = useModal();
    const t = useTranslations("profile.application");
    const router = useRouter();
    const queryClient = useQueryClient();
    const deleteDisplayApp = (appId: number) => deleteApp(locale, appId).then(() => {
        removeQuery(queryClient, ["applications"], appId);
    });

    useEffect(() => {
        if (data)
            // eslint-disable-next-line react-hooks/set-state-in-effect
            setApps(data.data);
    }, [data]);

    return (<div className="mx-auto space-y-6 text-center">
        {
            (apps && apps.length > 0) ?
            <Pagination currenIndex={index} totalPage={totalPage} onClick={setIndex} variableMT={true}>
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
                    {apps.map((a, idx) => <Application locale={locale} pageIndex={index} openModal={openModal} deleteDisplayApp={deleteDisplayApp} key={idx} app={a} t={t}/>)}
                </div>
            </Pagination> :
                <SmallText>{t("noApplicationsYet")}</SmallText>
        }
        <CreateNewApplication pageIndex={index} openModal={openModal} t={t} />
        <TextButton onClick={() => router.push("/api-documentation")}>{t("seeAPIDocumentation")}</TextButton>
    </div>);
}

function Application({app, openModal, deleteDisplayApp, locale, pageIndex, t}: {app: iApplication, openModal: (modal: ModalState) => void, deleteDisplayApp: (appId: number) => Promise<void>, locale: string, pageIndex: number, t: (txt: string) => string}) {
    const date = new Date(app.created_at);
    const createdAt = new Intl.DateTimeFormat(locale, {day: "2-digit", month: "2-digit", year: "numeric"}).format(date).replace(/[\/-]/g, ".");

    return (<div className="w-full flex justify-between group/card items-center border p-5 custom-shadow-animation-l text-left">
        <div className="space-y-1 w-full">
            <div className="flex justify-between items-center hover:cursor-pointer group/title mb-2">
                <div className="flex w-9/10 items-end gap-2" onClick={() => openModal({type: "application", appId: app.id, pageIndex: pageIndex})}>
                    <h3 className="group-hover/title:underline truncate">{app.name}</h3>
                    <span>{createdAt}</span>
                </div>
                <div className="opacity-0 group-hover/card:opacity-100">
                    <IconButton onClick={() => openModal({type: "delete-confirmation", deleteFunc: deleteDisplayApp, deleteObjId: app.id})} hoverColor="red">{(color: string) => <TrashIcon color={color} size={25}/>}</IconButton>
                </div>
            </div>
            <Code label="client_id">{app.client_id}</Code>
            {app.client_secret && <Code label="client_secret">{app.client_secret}</Code>}
            <div className="mt-2"><Label>{t("createModal.redirect_uri")}</Label>: <a href={app.redirect_uri} target="_blank" className="inline hover:underline hover:underline-offset-3 text-sm text-gray">{app.redirect_uri}</a></div>
        </div>
    </div>)
}

function CreateNewApplication({t, openModal, pageIndex}: {t: (str: string) => string, openModal: (modal: ModalState) => void, pageIndex: number}) {
    return (<div className="text-center">
        <Button onClick={() => openModal({type: "application", pageIndex: pageIndex})}>{t("createApplication")}</Button>
    </div>)
}
