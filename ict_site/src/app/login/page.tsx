"use client";

import axios from "axios";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { formatDateTime } from "@/lib/datetime";
import UIFormLogin from "@/uix/form/login";

type UserProfile = {
  id: string;
  username: string;
  email: string;
  fullname: string;
  phone?: string;
  role?: string;
  company_id: string;
  is_admin: boolean;
  is_hris: boolean;
  is_active: boolean;
};

type CompanyInfo = {
  id: string;
  name: string;
  slug: string;
};

type SessionData = {
  token: string;
  expires_at: string;
  user_profile: UserProfile;
};

type BackendLoginResponse = {
  message?: string;
  data: SessionData;
};

const loginSchema = z.object({
  company: z.string(),
  username: z.string().min(1, "Nama Pengguna wajib diisi"),
  password: z.string().min(1, "Kata Sandi wajib diisi"),
  captcha: z.string().min(1, "Test Keamanan wajib diisi"),
});

type LoginForm = z.infer<typeof loginSchema>;

const storageKey = "sessionMemorySave";

const parseSession = (value: string | null): SessionData | null => {
  if (!value) return null;
  try {
    const parsed = JSON.parse(value) as SessionData;
    if (!parsed?.token || !parsed?.expires_at) return null;
    return parsed;
  } catch {
    return null;
  }
};

const isSessionExpired = (session: SessionData | null) => {
  if (!session?.expires_at) return true;
  return new Date(session.expires_at).getTime() <= Date.now();
};

export default function LoginPage() {
  const router = useRouter();
  const [isReady, setIsReady] = useState(false);
  const [isBusy, setIsBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [captchaNum1, setCaptchaNum1] = useState(0);
  const [captchaNum2, setCaptchaNum2] = useState(0);
  const [todayDate, setTodayDate] = useState("");
  const [companies, setCompanies] = useState<CompanyInfo[]>([]);

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { company: "", username: "", password: "", captcha: "" },
  });

  const generateCaptcha = () => {
    const n1 = Math.floor(Math.random() * 10);
    const n2 = Math.floor(Math.random() * 10);
    setCaptchaNum1(n1);
    setCaptchaNum2(n2);
    form.setValue("captcha", "");
  };

  useEffect(() => {
    const stored = parseSession(window.localStorage.getItem(storageKey));
    if (stored && !isSessionExpired(stored)) {
      router.replace("/board");
    } else {
      window.localStorage.removeItem(storageKey);
      void axios
        .get("/api/data/companies")
        .then((res) => {
          const companyList: CompanyInfo[] = res.data?.data ?? [];
          setCompanies([{ id: "", name: "- Multi Company -", slug: "" }, ...companyList]);
        })
        .catch(() => {
          setCompanies([{ id: "", name: "- Multi Company -", slug: "" }]);
        });
      const today = formatDateTime(new Date());
      setTodayDate(today);
      generateCaptcha();
      setIsReady(true);
    }
  }, [router]);

  const onSubmit = async (values: LoginForm) => {
    const expected = String(captchaNum1 + captchaNum2);
    if (values.captcha !== expected) {
      setError("Test Keamanan tidak sesuai.");
      generateCaptcha();
      return;
    }

    setError(null);
    setMessage(null);
    setIsBusy(true);

    try {
      const response = await axios.post<BackendLoginResponse>("/api/auth/login", values, {
        headers: { "Content-Type": "application/json" },
      });

      const data = response.data.data ?? response.data;
      window.localStorage.setItem(storageKey, JSON.stringify(data));
      router.push("/board");
    } catch (err) {
      if (axios.isAxiosError(err) && err.response) {
        setError(err.response.data?.error || "Masuk sistem gagal. Periksa nama pengguna dan kata sandi.");
      } else {
        setError("Terjadi kesalahan jaringan saat masuk sistem.");
      }
      generateCaptcha();
    } finally {
      setIsBusy(false);
    }
  };

  if (!isReady) {
    return <div className="min-h-screen bg-slate-50" />;
  }

  return (
    <UIFormLogin
      form={form}
      onSubmit={form.handleSubmit(onSubmit)}
      companies={companies}
      captchaNum1={captchaNum1}
      captchaNum2={captchaNum2}
      generateCaptcha={generateCaptcha}
      isBusy={isBusy}
      message={message}
      error={error}
      todayDate={todayDate}
    />
  );
}
