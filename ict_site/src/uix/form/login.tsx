"use client";

import React from "react";
import { UseFormReturn } from "react-hook-form";

type CompanyInfo = {
  id: string;
  name: string;
  slug: string;
};

interface iFormLogin {
  form: UseFormReturn<any>;
  onSubmit: (e: React.FormEvent<HTMLFormElement>) => void;
  companies: CompanyInfo[];
  captchaNum1: number;
  captchaNum2: number;
  generateCaptcha: () => void;
  isBusy: boolean;
  message: string | null;
  error: string | null;
  todayDate: string;
}

export default function FormLogin({
  form,
  onSubmit,
  companies,
  captchaNum1,
  captchaNum2,
  generateCaptcha,
  isBusy,
  message,
  error,
  todayDate,
}: iFormLogin) {
  return (
    <div className="min-h-screen bg-slate-50 px-4 py-10 text-slate-900">
      <div className="mx-auto w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 shadow-xl shadow-slate-200/50">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-semibold text-slate-900">XBoard</h1>
        </div>

        {message && (
          <div className="mb-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-emerald-900">
            {message}
          </div>
        )}

        {error && (
          <div className="mb-4 rounded-xl border border-rose-200 bg-rose-50 p-4 text-rose-900">
            {error}
          </div>
        )}

        <form onSubmit={onSubmit} className="grid gap-4">
          <label className="grid gap-2 text-sm font-medium text-slate-700">
            Perusahaan
            <select
              {...form.register("company")}
              className="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm outline-none transition focus:border-slate-400 focus:ring-2 focus:ring-slate-100"
            >
              {companies.map((company) => (
                <option key={company.id || "empty-company"} value={company.id}>
                  {company.name}
                </option>
              ))}
            </select>
          </label>

          <label className="grid gap-2 text-sm font-medium text-slate-700">
            Nama Pengguna
            <input
              type="text"
              autoComplete="username"
              {...form.register("username")}
              className="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm outline-none transition focus:border-slate-400 focus:ring-2 focus:ring-slate-100"
            />
            {form.formState.errors.username && (
              <span className="text-sm text-rose-600">
                {form.formState.errors.username.message as string}
              </span>
            )}
          </label>

          <label className="grid gap-2 text-sm font-medium text-slate-700">
            Kata Sandi
            <input
              type="password"
              autoComplete="current-password"
              {...form.register("password")}
              className="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm outline-none transition focus:border-slate-400 focus:ring-2 focus:ring-slate-100"
            />
            {form.formState.errors.password && (
              <span className="text-sm text-rose-600">
                {form.formState.errors.password.message as string}
              </span>
            )}
          </label>

          <label className="grid gap-2 text-sm font-medium text-slate-700">
            <div className="flex items-center justify-between">
              <span>Test Keamanan : {captchaNum1} + {captchaNum2} = ?</span>
              <button
                type="button"
                onClick={generateCaptcha}
                className="text-xs text-slate-500 hover:text-slate-700 underline"
              >
                Refresh
              </button>
            </div>
            <input
              type="text"
              inputMode="numeric"
              {...form.register("captcha")}
              className="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm outline-none transition focus:border-slate-400 focus:ring-2 focus:ring-slate-100"
              placeholder="Jawab pertanyaan"
            />
            {form.formState.errors.captcha && (
              <span className="text-sm text-rose-600">
                {form.formState.errors.captcha.message as string}
              </span>
            )}
          </label>

          <button
            type="submit"
            disabled={isBusy}
            className="inline-flex h-12 items-center justify-center rounded-xl bg-slate-900 px-5 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isBusy ? "Memproses..." : "Masuk Sistem"}
          </button>
        </form>

        <div className="mt-8 text-center">
          <p className="text-xs text-slate-500">{todayDate}</p>
        </div>
      </div>
    </div>
  );
}
