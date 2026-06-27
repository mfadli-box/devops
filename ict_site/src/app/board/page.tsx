"use client";

import axios from "axios";
import { useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { formatDateTime } from "@/lib/datetime";

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

type SessionData = {
  token: string;
  expires_at: string;
  user_profile: UserProfile;
};

type CompanyInfo = {
  id: string;
  name: string;
  slug: string;
};

type ModuleNode = {
  id: string;
  code: string;
  title: string;
  path: string;
  is_page: boolean;
  level: "hide" | "view" | "book" | "post";
  children?: ModuleNode[];
};

type BoardData = {
  companies: CompanyInfo[];
  modules: ModuleNode[];
};

type UserAction = {
  id: string;
  action: string;
  module_code: string;
  path: string;
  ip_address: string;
  created_at: string;
};

type ProfileTab = "info" | "profile" | "password" | "history";
type MainView = "workspace" | "admin";
type AdminTab = "company" | "module" | "user";

type AdminCompany = {
  id: string;
  slug: string;
  name: string;
  vat_id: string;
  reg_no: string;
  address: string;
  valuta: string;
  hris_link: string;
  is_active: boolean;
};

type AdminModule = {
  id: string;
  parent_id: string;
  code: string;
  name: string;
  path: string;
  is_page: boolean;
  is_active: boolean;
};

type AdminUser = {
  id: string;
  company_id: string;
  username: string;
  email: string;
  fullname: string;
  phone: string;
  role: string;
  password?: string;
  is_admin: boolean;
  is_hris: boolean;
  is_active: boolean;
};

type AdminCompanyModuleAccess = {
  id: string;
  company_id: string;
  module_id: string;
  module_code: string;
  module_name: string;
  is_active: boolean;
};

type AdminUserCompanyAccess = {
  id: string;
  user_id: string;
  company_id: string;
  company_name: string;
  is_active: boolean;
};

type AdminUserPrivilegeAccess = {
  id: string;
  user_company_id: string;
  module_id: string;
  module_code: string;
  module_name: string;
  level: "hide" | "view" | "book" | "post";
};

type AlertTone = "error" | "success";

function DismissibleAlert({
  tone,
  message,
  onClose,
  className = "mt-4",
}: {
  tone: AlertTone;
  message: string;
  onClose: () => void;
  className?: string;
}) {
  const toneClassName =
    tone === "error"
      ? "border-rose-200 bg-rose-50 text-rose-700"
      : "border-emerald-200 bg-emerald-50 text-emerald-700";

  return (
    <div className={`${className} flex items-start justify-between gap-3 rounded-xl border p-3 text-sm ${toneClassName}`}>
      <span>{message}</span>
      <button
        type="button"
        onClick={onClose}
        aria-label="Tutup notifikasi"
        className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-current/20 hover:bg-white/40"
      >
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
          <path d="M18 6 6 18" />
          <path d="m6 6 12 12" />
        </svg>
      </button>
    </div>
  );
}

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

const normalizeBool = (value: unknown, fallback = false): boolean => {
  if (typeof value === "boolean") return value;
  if (typeof value === "number") return value !== 0;
  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    if (["true", "1", "t", "yes", "y", "on", "active"].includes(normalized)) return true;
    if (["false", "0", "f", "no", "n", "off", "inactive", ""].includes(normalized)) return false;
  }
  return fallback;
};

const isSessionExpired = (session: SessionData | null) => {
  if (!session?.expires_at) return true;
  return new Date(session.expires_at).getTime() <= Date.now();
};

const renderModuleTree = (
  nodes: ModuleNode[],
  level = 0,
  selectedModule: string | null,
  onSelect: (moduleId: string, isPage: boolean) => void
) => {
  return (
    <ul className={`space-y-0 ${level > 0 ? "pl-4" : ""}`}>
      {nodes.map((node) => (
        <li key={node.id}>
          <button
            type="button"
            onClick={() => onSelect(node.id, node.is_page)}
            className={`flex w-full items-center justify-between px-2 py-2 text-left transition ${selectedModule === node.id ? "bg-slate-900 text-white" : "bg-slate-50 text-slate-700 hover:bg-slate-100"}`}
          >
            <span>{node.title}</span>
            {node.children ? <span className="text-xs text-slate-400">+</span> : null}
          </button>
          {node.children ? renderModuleTree(node.children, level + 1, selectedModule, onSelect) : null}
        </li>
      ))}
    </ul>
  );
};

export default function BoardPage() {
  const router = useRouter();
  const [session, setSession] = useState<SessionData | null>(null);
  const [selectedCompany, setSelectedCompany] = useState<string>("");
  const [selectedModule, setSelectedModule] = useState<string | null>(null);
  const [boardData, setBoardData] = useState<BoardData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isBusy, setIsBusy] = useState(false);
  const [mn01PostForm, setMn01PostForm] = useState({
    title: "",
    note: "",
  });
  const [mn01PostRows, setMn01PostRows] = useState<Array<{ id: string; title: string; note: string; createdAt: string }>>([]);

  const [showProfileModal, setShowProfileModal] = useState(false);
  const [activeProfileTab, setActiveProfileTab] = useState<ProfileTab>("info");
  const [activeMainView, setActiveMainView] = useState<MainView>("workspace");
  const [activeAdminTab, setActiveAdminTab] = useState<AdminTab>("company");
  const [profileData, setProfileData] = useState<UserProfile | null>(null);
  const [historyData, setHistoryData] = useState<UserAction[]>([]);
  const [profileLoading, setProfileLoading] = useState(false);
  const [profileBusy, setProfileBusy] = useState(false);
  const [profileError, setProfileError] = useState<string | null>(null);
  const [profileMessage, setProfileMessage] = useState<string | null>(null);

  const [profileForm, setProfileForm] = useState({
    fullname: "",
    email: "",
    phone: "",
  });

  const [passwordForm, setPasswordForm] = useState({
    current_password: "",
    new_password: "",
    confirm_password: "",
  });
  const [adminLoading, setAdminLoading] = useState(false);
  const [adminBusy, setAdminBusy] = useState(false);
  const [adminError, setAdminError] = useState<string | null>(null);
  const [adminMessage, setAdminMessage] = useState<string | null>(null);
  const [adminFormError, setAdminFormError] = useState<string | null>(null);
  const [adminCompanies, setAdminCompanies] = useState<AdminCompany[]>([]);
  const [adminModules, setAdminModules] = useState<AdminModule[]>([]);
  const [adminUsers, setAdminUsers] = useState<AdminUser[]>([]);
  const [companyFilter, setCompanyFilter] = useState("");
  const [moduleFilter, setModuleFilter] = useState("");
  const [userFilter, setUserFilter] = useState("");
  const [companyPage, setCompanyPage] = useState(1);
  const [modulePage, setModulePage] = useState(1);
  const [userPage, setUserPage] = useState(1);
  const [historyFilter, setHistoryFilter] = useState("");
  const [historyPage, setHistoryPage] = useState(1);
  const [showCompanyDialog, setShowCompanyDialog] = useState(false);
  const [showModuleDialog, setShowModuleDialog] = useState(false);
  const [showUserDialog, setShowUserDialog] = useState(false);
  const [isMobileSidebar, setIsMobileSidebar] = useState(false);
  const [isCompactHeader, setIsCompactHeader] = useState(false);
  const [isModuleSidebarOpen, setIsModuleSidebarOpen] = useState(false);
  const [companyDialogTab, setCompanyDialogTab] = useState<"profile" | "access">("profile");
  const [userDialogTab, setUserDialogTab] = useState<"profile" | "companyAccess" | "moduleAccess">("profile");
  const [companyModuleAccessRows, setCompanyModuleAccessRows] = useState<AdminCompanyModuleAccess[]>([]);
  const [companyModuleAccessForm, setCompanyModuleAccessForm] = useState({ id: "", module_id: "", is_active: true });
  const [userCompanyAccessRows, setUserCompanyAccessRows] = useState<AdminUserCompanyAccess[]>([]);
  const [userCompanyAccessForm, setUserCompanyAccessForm] = useState({ id: "", company_id: "", is_active: true });
  const [selectedUserCompanyAccess, setSelectedUserCompanyAccess] = useState("");
  const [userPrivilegeAccessRows, setUserPrivilegeAccessRows] = useState<AdminUserPrivilegeAccess[]>([]);
  const [userPrivilegeAccessForm, setUserPrivilegeAccessForm] = useState({ id: "", user_company_id: "", module_id: "", level: "hide" as "hide" | "view" | "book" | "post" });
  const userPrivilegeAccessFormLabels: Record<string, string> = {
    hide: "Sembunyi",
    view: "Lihat",
    book: "Pesan",
    post: "Proses",
  };
  const [companyAdminForm, setCompanyAdminForm] = useState<AdminCompany>({
    id: "",
    slug: "",
    name: "",
    vat_id: "",
    reg_no: "",
    address: "",
    valuta: "IDR",
    hris_link: "",
    is_active: true,
  });
  const [moduleAdminForm, setModuleAdminForm] = useState<AdminModule>({
    id: "",
    parent_id: "",
    code: "",
    name: "",
    path: "",
    is_page: true,
    is_active: true,
  });
  const [userAdminForm, setUserAdminForm] = useState<AdminUser>({
    id: "",
    company_id: "",
    username: "",
    email: "",
    fullname: "",
    phone: "",
    role: "staff",
    password: "",
    is_admin: false,
    is_hris: false,
    is_active: true,
  });
  const profileCompanyID = (userAdminForm.company_id || "").trim();
  const isProfileCompanyLocked = Boolean(userAdminForm.id && profileCompanyID !== "");
  const hasProfileCompanyMapping = !isProfileCompanyLocked || Boolean(selectedUserCompanyAccess);

  const forceClientLogout = () => {
    window.localStorage.removeItem(storageKey);
    try {
      document.cookie = "sessionMemorySave=; path=/; max-age=0";
    } catch (e) {}
    router.replace("/login");
  };

  useEffect(() => {
    const stored = parseSession(window.localStorage.getItem(storageKey));
    if (!stored || isSessionExpired(stored)) {
      window.localStorage.removeItem(storageKey);
      router.replace("/login");
      return;
    }
    setSession(stored);
    setSelectedCompany(stored.user_profile.company_id ?? "");
  }, [router]);

  useEffect(() => {
    if (!session) return;

    let mounted = true;
    const headers = { Authorization: `Bearer ${session.token}` };

    const fetchCompanies = async () => {
      try {
        const cmpRes = await axios.get("/api/data/companies", { headers });
        if (!mounted) return;
        const companies: CompanyInfo[] = [{ id: "", name: "- Pilih Perusahaan -", slug: "" }, ...(cmpRes.data?.data ?? [])];
        setBoardData((prev) => ({ companies, modules: prev?.modules ?? [] }));

        const companyExists = companies.some((item) => item.id !== "" && item.id === selectedCompany);
        if (!companyExists) {
          const fallbackCompany = companies.find((item) => item.id !== "")?.id ?? "";
          setSelectedCompany(fallbackCompany);
          setSelectedModule(null);
        }
      } catch (err) {
        console.error("Gagal mengambil data perusahaan", err);
        setBoardData({ companies: [{ id: "", name: "- Pilih Perusahaan -", slug: "" }], modules: [] });
      }
    };
    fetchCompanies();
    return () => {
      mounted = false;
    };
  }, [session]);

  useEffect(() => {
    if (!session) return;

    let mounted = true;
    const headers = { Authorization: `Bearer ${session.token}` };

    const fetchModules = async () => {
      if (!selectedCompany) {
        if (!mounted) return;
        setBoardData((prev) => ({ companies: prev?.companies ?? [{ id: "", name: "- Pilih Perusahaan -", slug: "" }], modules: [] }));
        setSelectedModule(null);
        return;
      }

      try {
        const modRes = await axios.get("/api/data/modules", { headers, params: { company_id: selectedCompany } });
        if (!mounted) return;
        const modules: ModuleNode[] = modRes.data?.data ?? [];
        setBoardData((prev) => ({ companies: prev?.companies ?? [{ id: "", name: "- Pilih Perusahaan -", slug: "" }], modules }));

        const hasSelected = (() => {
          if (!selectedModule) return true;
          const walk = (nodes: ModuleNode[]): boolean =>
            nodes.some((node) => node.id === selectedModule || (node.children ? walk(node.children) : false));
          return walk(modules);
        })();

        if (!hasSelected) {
          setSelectedModule(null);
        }
      } catch (err) {
        console.error("Gagal mengambil data modul", err);
        if (!mounted) return;
        setBoardData((prev) => ({ companies: prev?.companies ?? [{ id: "", name: "- Pilih Perusahaan -", slug: "" }], modules: [] }));
        setSelectedModule(null);
      }
    };

    fetchModules();
    return () => {
      mounted = false;
    };
  }, [session, selectedCompany]);

  const company = useMemo(() => {
    const list = boardData?.companies ?? [];
    if (!selectedCompany) {
      return { id: "", name: "", slug: "" };
    }
    return list.find((item) => item.id === selectedCompany) || { id: "", name: "", slug: "" };
  }, [selectedCompany, boardData]);

  const selectedModuleNode = useMemo(() => {
    const findNode = (nodes: ModuleNode[]): ModuleNode | null => {
      for (const node of nodes) {
        if (node.id === selectedModule) return node;
        if (node.children) {
          const found = findNode(node.children);
          if (found) return found;
        }
      }
      return null;
    };
    return selectedModule ? findNode(boardData?.modules ?? []) : null;
  }, [selectedModule, boardData]);

  const isMN01Module = useMemo(() => {
    if (!selectedModuleNode) return false;
    const code = (selectedModuleNode.code || "").trim().toUpperCase();
    const path = (selectedModuleNode.path || "").trim().toUpperCase();
    return code === "MN01" || path === "MN01";
  }, [selectedModuleNode]);

  const moduleBreadcrumb = useMemo(() => {
    const path: ModuleNode[] = [];
    const findPath = (nodes: ModuleNode[], target: string): boolean => {
      for (const node of nodes) {
        if (node.id === target) {
          path.push(node);
          return true;
        }
        if (node.children) {
          if (findPath(node.children, target)) {
            path.unshift(node);
            return true;
          }
        }
      }
      return false;
    };
    if (selectedModule) {
      findPath(boardData?.modules ?? [], selectedModule);
    }
    return path;
  }, [selectedModule, boardData]);

  const refreshProfileModalData = async () => {
    if (!session) return;
    setProfileLoading(true);
    setProfileError(null);

    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const [profileResult, historyResult] = await Promise.allSettled([
        axios.get("/api/user/profile", { headers }),
        axios.get("/api/user/history", { headers }),
      ]);

      if (profileResult.status === "fulfilled") {
        const profile = profileResult.value.data?.data ?? null;
        setProfileData(profile);

        if (profile) {
          setProfileForm({
            fullname: profile.fullname ?? "",
            email: profile.email ?? "",
            phone: profile.phone ?? "",
          });
        }
      } else {
        const err = profileResult.reason;
        if (axios.isAxiosError(err) && (err.response?.status === 401 || err.response?.status === 403)) {
          forceClientLogout();
          return;
        }
        setProfileError(
          axios.isAxiosError(err)
            ? err.response?.data?.error || "Gagal memuat data profil."
            : "Gagal memuat data profil."
        );
      }

      if (historyResult.status === "fulfilled") {
        const history = historyResult.value.data?.data ?? [];
        setHistoryData(Array.isArray(history) ? history : []);
      } else {
        setHistoryData([]);
      }

      if (!profileData && profileResult.status === "fulfilled") {
        const profile = profileResult.value.data?.data ?? null;
        if (profile) {
          setProfileData(profile);
        }
      }

      if (profileResult.status === "fulfilled" && profileResult.value.data?.data) {
        const profile = profileResult.value.data.data;
        setProfileForm({
          fullname: profile.fullname ?? "",
          email: profile.email ?? "",
          phone: profile.phone ?? "",
        });
      }

      if (profileResult.status === "fulfilled" && historyResult.status === "rejected") {
        setProfileMessage("Data riwayat belum tersedia saat ini.");
      }
    } catch {
      setProfileError("Gagal memuat data profil.");
    } finally {
      setProfileLoading(false);
    }
  };

  const openProfileModal = async () => {
    setShowProfileModal(true);
    setActiveProfileTab("info");
    setProfileMessage(null);
    setProfileError(null);
    setHistoryFilter("");
    setHistoryPage(1);
    await refreshProfileModalData();
  };

  const resetAdminForms = () => {
    setAdminFormError(null);
    setCompanyDialogTab("profile");
    setUserDialogTab("profile");
    setCompanyAdminForm({ id: "", slug: "", name: "", vat_id: "", reg_no: "", address: "", valuta: "IDR", hris_link: "", is_active: true });
    setModuleAdminForm({ id: "", parent_id: "", code: "", name: "", path: "", is_page: true, is_active: true });
    setUserAdminForm({ id: "", company_id: "", username: "", email: "", fullname: "", phone: "", role: "staff", password: "", is_admin: false, is_hris: false, is_active: true });
    setCompanyModuleAccessRows([]);
    setCompanyModuleAccessForm({ id: "", module_id: "", is_active: true });
    setUserCompanyAccessRows([]);
    setUserCompanyAccessForm({ id: "", company_id: "", is_active: true });
    setSelectedUserCompanyAccess("");
    setUserPrivilegeAccessRows([]);
    setUserPrivilegeAccessForm({ id: "", user_company_id: "", module_id: "", level: "hide" });
  };

  const pageSize = 10;
  const historyPageSize = 5;

  const historyFiltered = useMemo(() => {
    const query = historyFilter.trim().toLowerCase();
    if (!query) return historyData;

    return historyData.filter((item) =>
      [formatDateTime(item.created_at), item.action, item.module_code, item.path]
        .join(" ")
        .toLowerCase()
        .includes(query)
    );
  }, [historyData, historyFilter]);

  const historyPageTotal = Math.max(1, Math.ceil(historyFiltered.length / historyPageSize));
  const historyRows = historyFiltered.slice((historyPage - 1) * historyPageSize, historyPage * historyPageSize);

  const companyFiltered = useMemo(() => {
    const q = companyFilter.trim().toLowerCase();
    if (!q) return adminCompanies;
    return adminCompanies.filter((item) =>
      [item.name, item.slug, item.valuta].join(" ").toLowerCase().includes(q)
    );
  }, [adminCompanies, companyFilter]);

  const moduleFiltered = useMemo(() => {
    const q = moduleFilter.trim().toLowerCase();
    if (!q) return adminModules;
    return adminModules.filter((item) =>
      [item.code, item.name, item.path].join(" ").toLowerCase().includes(q)
    );
  }, [adminModules, moduleFilter]);

  const userFiltered = useMemo(() => {
    const q = userFilter.trim().toLowerCase();
    if (!q) return adminUsers;
    return adminUsers.filter((item) =>
      [item.username, item.email, item.fullname, item.role].join(" ").toLowerCase().includes(q)
    );
  }, [adminUsers, userFilter]);

  const companyModuleInputOptions = useMemo(() => {
    const selectedOnForm = companyModuleAccessForm.module_id;
    const usedModuleIDs = new Set(companyModuleAccessRows.map((item) => item.module_id));
    return adminModules.filter((item) => {
      if (!item.is_page) return false;
      if (!companyModuleAccessForm.id && usedModuleIDs.has(item.id)) return false;
      if (companyModuleAccessForm.id && item.id !== selectedOnForm && usedModuleIDs.has(item.id)) return false;
      return true;
    });
  }, [adminModules, companyModuleAccessRows, companyModuleAccessForm.id, companyModuleAccessForm.module_id]);

  const userCompanyInputOptions = useMemo(() => {
    const selectedOnForm = userCompanyAccessForm.company_id;
    const usedCompanyIDs = new Set(userCompanyAccessRows.map((item) => item.company_id));
    return adminCompanies.filter((item) => {
      if (!userCompanyAccessForm.id && usedCompanyIDs.has(item.id)) return false;
      if (userCompanyAccessForm.id && item.id !== selectedOnForm && usedCompanyIDs.has(item.id)) return false;
      return true;
    });
  }, [adminCompanies, userCompanyAccessRows, userCompanyAccessForm.id, userCompanyAccessForm.company_id]);

  const selectedUserCompanyRow = useMemo(
    () => userCompanyAccessRows.find((item) => item.id === selectedUserCompanyAccess),
    [userCompanyAccessRows, selectedUserCompanyAccess]
  );

  const selectedPrivilegeCompany = useMemo(
    () => adminCompanies.find((item) => item.id === (selectedUserCompanyRow?.company_id || "")),
    [adminCompanies, selectedUserCompanyRow]
  );

  const userPrivilegeModuleInputOptions = useMemo(() => {
    if (!selectedUserCompanyRow) return [];
    const activeCompanyModuleIDs = new Set(
      companyModuleAccessRows.filter((item) => item.is_active).map((item) => item.module_id)
    );
    const usedPrivilegeModuleIDs = new Set(userPrivilegeAccessRows.map((item) => item.module_id));
    const selectedOnForm = userPrivilegeAccessForm.module_id;

    return adminModules.filter((item) => {
      if (!item.is_active || !item.is_page) return false;
      if (!activeCompanyModuleIDs.has(item.id)) return false;
      if (!userPrivilegeAccessForm.id && usedPrivilegeModuleIDs.has(item.id)) return false;
      if (userPrivilegeAccessForm.id && item.id !== selectedOnForm && usedPrivilegeModuleIDs.has(item.id)) return false;
      return true;
    });
  }, [adminModules, companyModuleAccessRows, selectedUserCompanyRow, userPrivilegeAccessRows, userPrivilegeAccessForm.id, userPrivilegeAccessForm.module_id]);

  const companyPageTotal = Math.max(1, Math.ceil(companyFiltered.length / pageSize));
  const modulePageTotal = Math.max(1, Math.ceil(moduleFiltered.length / pageSize));
  const userPageTotal = Math.max(1, Math.ceil(userFiltered.length / pageSize));

  const companyRows = companyFiltered.slice((companyPage - 1) * pageSize, companyPage * pageSize);
  const moduleRows = moduleFiltered.slice((modulePage - 1) * pageSize, modulePage * pageSize);
  const userRows = userFiltered.slice((userPage - 1) * pageSize, userPage * pageSize);

  useEffect(() => {
    setCompanyPage(1);
  }, [companyFilter]);

  useEffect(() => {
    if (companyPage > companyPageTotal) {
      setCompanyPage(companyPageTotal);
    }
  }, [companyPage, companyPageTotal]);

  useEffect(() => {
    setModulePage(1);
  }, [moduleFilter]);

  useEffect(() => {
    if (modulePage > modulePageTotal) {
      setModulePage(modulePageTotal);
    }
  }, [modulePage, modulePageTotal]);

  useEffect(() => {
    setUserPage(1);
  }, [userFilter]);

  useEffect(() => {
    if (userPage > userPageTotal) {
      setUserPage(userPageTotal);
    }
  }, [userPage, userPageTotal]);

  useEffect(() => {
    setHistoryPage(1);
  }, [historyFilter]);

  useEffect(() => {
    if (historyPage > historyPageTotal) {
      setHistoryPage(historyPageTotal);
    }
  }, [historyPage, historyPageTotal]);

  useEffect(() => {
    const syncViewport = () => {
      const mobile = window.innerWidth < 1024;
      const compactHeader = window.innerWidth < 800;
      setIsMobileSidebar(mobile);
      setIsCompactHeader(compactHeader);
      setIsModuleSidebarOpen((current) => (mobile ? current : true));
    };

    syncViewport();
    window.addEventListener("resize", syncViewport);
    return () => window.removeEventListener("resize", syncViewport);
  }, []);

  const fetchAdminData = async () => {
    if (!session?.user_profile.is_admin || !session) return;
    setAdminLoading(true);
    setAdminError(null);
    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const [companiesRes, modulesRes, usersRes] = await Promise.all([
        axios.get("/api/admin/company", { headers }),
        axios.get("/api/admin/module", { headers }),
        axios.get("/api/admin/user", { headers }),
      ]);
      setAdminCompanies(companiesRes.data?.data ?? []);
      setAdminModules(modulesRes.data?.data ?? []);
      setAdminUsers(usersRes.data?.data ?? []);
    } catch (err) {
      if (axios.isAxiosError(err) && (err.response?.status === 401 || err.response?.status === 403)) {
        if (err.response?.status === 401) {
          forceClientLogout();
          return;
        }
        setAdminError("Akses admin ditolak.");
        return;
      }
      setAdminError("Gagal memuat menu admin.");
    } finally {
      setAdminLoading(false);
    }
  };

  const openAdminView = async () => {
    setActiveMainView("admin");
    setAdminMessage(null);
    setAdminError(null);
    setAdminFormError(null);
    resetAdminForms();
    if (isMobileSidebar) {
      setIsModuleSidebarOpen(false);
    }
    await fetchAdminData();
  };

  const saveAdminCompany = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session) return;
    setAdminBusy(true);
    setAdminError(null);
    setAdminMessage(null);
    setAdminFormError(null);
    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const payload = { ...companyAdminForm };
      if (companyAdminForm.id) {
        await axios.put("/api/admin/company", payload, { headers });
        setAdminMessage("Perusahaan berhasil diperbarui.");
      } else {
        await axios.post("/api/admin/company", payload, { headers });
        setAdminMessage("Perusahaan berhasil ditambahkan.");
      }
      setShowCompanyDialog(false);
      resetAdminForms();
      await fetchAdminData();
    } catch (err) {
      setAdminFormError(axios.isAxiosError(err) ? err.response?.data?.error || "Gagal menyimpan perusahaan." : "Gagal menyimpan perusahaan.");
    } finally {
      setAdminBusy(false);
    }
  };

  const saveAdminModule = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session) return;
    setAdminBusy(true);
    setAdminError(null);
    setAdminMessage(null);
    setAdminFormError(null);
    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const payload = { ...moduleAdminForm };
      if (moduleAdminForm.id) {
        await axios.put("/api/admin/module", payload, { headers });
        setAdminMessage("Modul berhasil diperbarui.");
      } else {
        await axios.post("/api/admin/module", payload, { headers });
        setAdminMessage("Modul berhasil ditambahkan.");
      }
      setShowModuleDialog(false);
      resetAdminForms();
      await fetchAdminData();
    } catch (err) {
      setAdminFormError(axios.isAxiosError(err) ? err.response?.data?.error || "Gagal menyimpan modul." : "Gagal menyimpan modul.");
    } finally {
      setAdminBusy(false);
    }
  };

  const saveAdminUser = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session) return;
    setAdminBusy(true);
    setAdminError(null);
    setAdminMessage(null);
    setAdminFormError(null);
    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const payload = { ...userAdminForm };
      if (userAdminForm.id) {
        await axios.put("/api/admin/user", payload, { headers });
        setAdminMessage("Pengguna berhasil diperbarui.");
      } else {
        await axios.post("/api/admin/user", payload, { headers });
        setAdminMessage("Pengguna berhasil ditambahkan.");
      }
      setShowUserDialog(false);
      resetAdminForms();
      await fetchAdminData();
    } catch (err) {
      setAdminFormError(axios.isAxiosError(err) ? err.response?.data?.error || "Gagal menyimpan pengguna." : "Gagal menyimpan pengguna.");
    } finally {
      setAdminBusy(false);
    }
  };

  const fetchCompanyModuleAccess = async (companyID: string) => {
    if (!session || !companyID) {
      setCompanyModuleAccessRows([]);
      return;
    }
    const headers = { Authorization: `Bearer ${session.token}` };
    const res = await axios.get("/api/admin/company-module", { headers, params: { company_id: companyID } });
    const rows = Array.isArray(res.data?.data)
      ? res.data.data.map((item: Partial<AdminCompanyModuleAccess>) => ({
          id: item.id ?? "",
          company_id: item.company_id ?? "",
          module_id: item.module_id ?? "",
          module_code: item.module_code ?? "",
          module_name: item.module_name ?? "",
          is_active: normalizeBool(item.is_active),
        }))
      : [];
    setCompanyModuleAccessRows(rows);
  };

  const saveCompanyModuleAccess = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session || !companyAdminForm.id) return;
    setAdminBusy(true);
    setAdminFormError(null);
    setAdminMessage(null);
    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const payload = {
        id: companyModuleAccessForm.id,
        company_id: companyAdminForm.id,
        module_id: companyModuleAccessForm.module_id,
        is_active: normalizeBool(companyModuleAccessForm.is_active),
      };
      if (companyModuleAccessForm.id) {
        await axios.put("/api/admin/company-module", payload, { headers });
        setAdminMessage("Hak akses perusahaan berhasil diperbarui.");
      } else {
        await axios.post("/api/admin/company-module", payload, { headers });
        setAdminMessage("Hak akses perusahaan berhasil ditambahkan.");
      }
      setCompanyModuleAccessForm({ id: "", module_id: "", is_active: true });
      await fetchCompanyModuleAccess(companyAdminForm.id);
    } catch (err) {
      setAdminFormError(axios.isAxiosError(err) ? err.response?.data?.error || "Gagal menyimpan hak akses perusahaan." : "Gagal menyimpan hak akses perusahaan.");
    } finally {
      setAdminBusy(false);
    }
  };

  const fetchUserCompanyAccess = async (userID: string) => {
    if (!session || !userID) {
      setUserCompanyAccessRows([]);
      setSelectedUserCompanyAccess("");
      setUserPrivilegeAccessRows([]);
      return;
    }

    const headers = { Authorization: `Bearer ${session.token}` };
    const res = await axios.get("/api/admin/user-company", { headers, params: { user_id: userID } });
    const rows: AdminUserCompanyAccess[] = Array.isArray(res.data?.data)
      ? res.data.data.map((item: Partial<AdminUserCompanyAccess>) => ({
          id: item.id ?? "",
          user_id: item.user_id ?? "",
          company_id: item.company_id ?? "",
          company_name: item.company_name ?? "",
          is_active: normalizeBool(item.is_active),
        }))
      : [];
    setUserCompanyAccessRows(rows);

    const profileMatchID = profileCompanyID ? rows.find((item) => item.company_id === profileCompanyID)?.id ?? "" : "";
    const activeUserCompanyID = isProfileCompanyLocked ? profileMatchID : (profileMatchID || rows[0]?.id || "");

    setSelectedUserCompanyAccess(activeUserCompanyID);
    setUserCompanyAccessForm({ id: "", company_id: profileCompanyID || "", is_active: true });
    setUserPrivilegeAccessForm({ id: "", user_company_id: activeUserCompanyID, module_id: "", level: "hide" });

    if (activeUserCompanyID) {
      await fetchUserPrivilegeAccess(activeUserCompanyID);
    } else {
      setUserPrivilegeAccessRows([]);
    }
  };

  const fetchUserPrivilegeAccess = async (userCompanyID: string) => {
    if (!session || !userCompanyID) {
      setUserPrivilegeAccessRows([]);
      return;
    }
    const headers = { Authorization: `Bearer ${session.token}` };
    const res = await axios.get("/api/admin/user-privilege", { headers, params: { user_company_id: userCompanyID } });
    setUserPrivilegeAccessRows(Array.isArray(res.data?.data) ? res.data.data : []);

    const selectedCompanyID = userCompanyAccessRows.find((item) => item.id === userCompanyID)?.company_id || "";
    if (selectedCompanyID) {
      const companyModuleRes = await axios.get("/api/admin/company-module", { headers, params: { company_id: selectedCompanyID } });
      const rows = Array.isArray(companyModuleRes.data?.data)
        ? companyModuleRes.data.data.map((item: Partial<AdminCompanyModuleAccess>) => ({
            id: item.id ?? "",
            company_id: item.company_id ?? "",
            module_id: item.module_id ?? "",
            module_code: item.module_code ?? "",
            module_name: item.module_name ?? "",
            is_active: normalizeBool(item.is_active),
          }))
        : [];
      setCompanyModuleAccessRows(rows);
    } else {
      setCompanyModuleAccessRows([]);
    }
  };

  const saveUserCompanyAccess = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session || !userAdminForm.id) return;
    if (isProfileCompanyLocked) {
      setAdminFormError("Akses Perusahaan dinonaktifkan karena profil pengguna sudah memiliki perusahaan tetap.");
      return;
    }
    setAdminBusy(true);
    setAdminFormError(null);
    setAdminMessage(null);
    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const payload = {
        id: userCompanyAccessForm.id,
        user_id: userAdminForm.id,
        company_id: userCompanyAccessForm.company_id,
        is_active: normalizeBool(userCompanyAccessForm.is_active),
      };
      if (userCompanyAccessForm.id) {
        await axios.put("/api/admin/user-company", payload, { headers });
        setAdminMessage("Akses perusahaan pengguna berhasil diperbarui.");
      } else {
        await axios.post("/api/admin/user-company", payload, { headers });
        setAdminMessage("Akses perusahaan pengguna berhasil ditambahkan.");
      }
      await fetchUserCompanyAccess(userAdminForm.id);
    } catch (err) {
      setAdminFormError(axios.isAxiosError(err) ? err.response?.data?.error || "Gagal menyimpan akses perusahaan pengguna." : "Gagal menyimpan akses perusahaan pengguna.");
    } finally {
      setAdminBusy(false);
    }
  };

  const saveUserPrivilegeAccess = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session || !userAdminForm.id) return;
    setAdminBusy(true);
    setAdminFormError(null);
    setAdminMessage(null);
    try {
      const headers = { Authorization: `Bearer ${session.token}` };
      const effectiveUserCompanyID = isProfileCompanyLocked
        ? selectedUserCompanyAccess
        : (userPrivilegeAccessForm.user_company_id || selectedUserCompanyAccess);
      if (!effectiveUserCompanyID) {
        setAdminFormError("Perusahaan pengguna tidak ditemukan untuk perusahaan pada profil pengguna.");
        return;
      }
      const payload = {
        id: userPrivilegeAccessForm.id,
        user_company_id: effectiveUserCompanyID,
        module_id: userPrivilegeAccessForm.module_id,
        level: userPrivilegeAccessForm.level,
      };
      if (userPrivilegeAccessForm.id) {
        await axios.put("/api/admin/user-privilege", payload, { headers });
        setAdminMessage("Hak akses modul pengguna berhasil diperbarui.");
      } else {
        await axios.post("/api/admin/user-privilege", payload, { headers });
        setAdminMessage("Hak akses modul pengguna berhasil ditambahkan.");
      }
      setUserPrivilegeAccessForm({ id: "", user_company_id: selectedUserCompanyAccess, module_id: "", level: "hide" });
      if (selectedUserCompanyAccess) {
        await fetchUserPrivilegeAccess(selectedUserCompanyAccess);
      }
    } catch (err) {
      setAdminFormError(axios.isAxiosError(err) ? err.response?.data?.error || "Gagal menyimpan hak akses modul pengguna." : "Gagal menyimpan hak akses modul pengguna.");
    } finally {
      setAdminBusy(false);
    }
  };

  const handleProfileUpdate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session) return;

    setProfileBusy(true);
    setProfileError(null);
    setProfileMessage(null);

    try {
      const res = await axios.put(
        "/api/user/profile",
        {
          fullname: profileForm.fullname,
          email: profileForm.email,
          phone: profileForm.phone,
        },
        { headers: { Authorization: `Bearer ${session.token}` } }
      );

      const updated = res.data?.data as UserProfile;
      if (updated) {
        setProfileData(updated);

        const current = parseSession(window.localStorage.getItem(storageKey));
        if (current) {
          const next: SessionData = {
            ...current,
            user_profile: {
              ...current.user_profile,
              fullname: updated.fullname,
              email: updated.email,
            },
          };
          window.localStorage.setItem(storageKey, JSON.stringify(next));
          setSession(next);
        }
      }

      setProfileMessage(res.data?.message || "Profil berhasil diperbarui.");
    } catch (err) {
      if (axios.isAxiosError(err) && (err.response?.status === 401 || err.response?.status === 403)) {
        forceClientLogout();
        return;
      }
      setProfileError(
        axios.isAxiosError(err)
          ? err.response?.data?.error || "Gagal memperbarui profil."
          : "Gagal memperbarui profil."
      );
    } finally {
      setProfileBusy(false);
    }
  };

  const handlePasswordUpdate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!session) return;

    setProfileBusy(true);
    setProfileError(null);
    setProfileMessage(null);

    try {
      const res = await axios.put("/api/user/password", passwordForm, {
        headers: { Authorization: `Bearer ${session.token}` },
      });
      setPasswordForm({ current_password: "", new_password: "", confirm_password: "" });
      setProfileMessage(res.data?.message || "Kata Sandi berhasil diperbarui.");
    } catch (err) {
      if (axios.isAxiosError(err) && (err.response?.status === 401 || err.response?.status === 403)) {
        if (err.response?.status === 401) {
          forceClientLogout();
          return;
        }
      }
      setProfileError(
        axios.isAxiosError(err)
          ? err.response?.data?.error || "Gagal memperbarui password."
          : "Gagal memperbarui password."
      );
    } finally {
      setProfileBusy(false);
    }
  };

  const handleLogout = async () => {
    if (!session) {
      forceClientLogout();
      return;
    }

    setError(null);
    setIsBusy(true);

    try {
      await axios.post(
        "/api/auth/logout",
        { sessionToken: session.token },
        {
          headers: {
            Authorization: `Bearer ${session.token}`,
            "Content-Type": "application/json",
          },
        }
      );
      forceClientLogout();
    } catch (err) {
      if (axios.isAxiosError(err) && err.response) {
        if (err.response.status === 401 || err.response.status === 403) {
          forceClientLogout();
          return;
        }
        setError(err.response.data?.error || "Keluar sistem gagal.");
      } else {
        setError("Terjadi kesalahan jaringan saat keluar sistem.");
      }
    } finally {
      setIsBusy(false);
    }
  };

  const handleCompanyChange = (companyId: string) => {
    setSelectedCompany(companyId);
    if (!companyId) {
      setSelectedModule(null);
      setActiveMainView("workspace");
    }
    if (isMobileSidebar) {
      setIsModuleSidebarOpen(false);
    }
  };

  const handleModuleSelect = (moduleId: string, isPage: boolean) => {
    if (!isPage) {
      return;
    }
    setSelectedModule(moduleId);
    setActiveMainView("workspace");
    if (isMobileSidebar) {
      setIsModuleSidebarOpen(false);
    }
  };

  const handleMN01Submit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!selectedModuleNode || selectedModuleNode.level !== "post") return;
    const title = mn01PostForm.title.trim();
    const note = mn01PostForm.note.trim();
    if (!title || !note) return;

    setMn01PostRows((prev) => [
      {
        id: `${Date.now()}`,
        title,
        note,
        createdAt: new Date().toISOString(),
      },
      ...prev,
    ]);

    setMn01PostForm({ title: "", note: "" });
  };

  useEffect(() => {
    if (!showUserDialog || userDialogTab !== "moduleAccess" || !selectedUserCompanyAccess) return;
    fetchUserPrivilegeAccess(selectedUserCompanyAccess).catch(() => {
      setAdminFormError("Gagal memuat user_privilege.");
    });
  }, [showUserDialog, userDialogTab, selectedUserCompanyAccess]);

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <div className="border-b border-slate-200 bg-white px-4 py-4 shadow-sm">
        <div className={`mx-auto flex max-w-8xl gap-4 ${isCompactHeader ? "flex-col items-stretch" : "items-center justify-between"}`}>
          <div className="flex-1">
            <h1 className="text-2xl font-semibold text-slate-900">{company.name || "---"}</h1>
            {moduleBreadcrumb.length > 0 ? (
              <div className="mt-2 flex items-center gap-2 text-sm text-slate-600">
                {moduleBreadcrumb.map((node, idx) => (
                  <div key={node.id} className="flex items-center gap-2">
                    <span>{node.title}</span>
                    {idx < moduleBreadcrumb.length - 1 && <span className="text-slate-400">/</span>}
                  </div>
                ))}
              </div>
            ) : (
              <p className="mt-2 text-sm text-slate-500">{session?.user_profile.fullname || "---"}</p>
            )}
          </div>
          <div className={`flex gap-2 ${isCompactHeader ? "flex-wrap" : ""}`}>
            <button
              type="button"
              onClick={() => setIsModuleSidebarOpen(true)}
              className="inline-flex h-10 w-10 items-center justify-center rounded-2xl border border-slate-200 bg-white text-slate-700 transition hover:bg-slate-50 lg:hidden"
              aria-label="Buka sidebar modul"
              title="Buka sidebar modul"
            >
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-5 w-5" aria-hidden="true">
                <path d="M3 6h18" />
                <path d="M3 12h18" />
                <path d="M3 18h18" />
              </svg>
            </button>
            {session?.user_profile.is_admin ? (
              <button
                type="button"
                onClick={openAdminView}
                className={`inline-flex h-10 items-center justify-center rounded-2xl px-4 text-sm font-medium transition ${activeMainView === "admin" ? "bg-slate-900 text-white" : "bg-amber-100 text-amber-800 hover:bg-amber-200"}`}
              >
                Admin
              </button>
            ) : null}
            <button
              type="button"
              onClick={openProfileModal}
              className="inline-flex h-10 items-center justify-center rounded-2xl bg-slate-100 px-4 text-sm font-medium text-slate-700 transition hover:bg-slate-200"
            >
              Profil
            </button>
            <button
              type="button"
              onClick={handleLogout}
              disabled={isBusy}
              className="inline-flex h-10 items-center justify-center rounded-2xl bg-rose-600 px-4 text-sm font-semibold text-white transition hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {isBusy ? "Keluar..." : "Keluar"}
            </button>
          </div>
        </div>
      </div>

      <div className="relative flex flex-1">
        {isMobileSidebar && isModuleSidebarOpen ? (
          <button
            type="button"
            aria-label="Tutup panel modul"
            onClick={() => setIsModuleSidebarOpen(false)}
            className="fixed inset-0 z-20 bg-black/30 lg:hidden"
          />
        ) : null}

        <div
          className={`border-r border-slate-200 bg-white ${isMobileSidebar ? "fixed inset-x-0 bottom-0 z-30 flex flex-col overflow-hidden shadow-xl transition-transform duration-300" : "w-80 shrink-0 overflow-y-auto px-4 py-4"} ${isMobileSidebar && !isModuleSidebarOpen ? "-translate-x-full" : "translate-x-0"}`}
          style={{
            top: isMobileSidebar ? (isCompactHeader ? 129 : 81) : undefined,
            height: isMobileSidebar ? `calc(100dvh - ${isCompactHeader ? 129 : 81}px)` : "calc(100vh - 81px)",
          }}
        >
          {isMobileSidebar ? (
            <div className="flex items-center justify-between border-b border-slate-200 px-4 py-4 lg:hidden">
              <p className="text-sm font-semibold text-slate-900">Sidebar Modul</p>
              <button
                type="button"
                aria-label="Tutup sidebar modul"
                title="Tutup"
                onClick={() => setIsModuleSidebarOpen(false)}
                className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                  <path d="M18 6 6 18" />
                  <path d="m6 6 12 12" />
                </svg>
              </button>
            </div>
          ) : null}
          <div className={`${isMobileSidebar ? "min-h-0 flex-1 space-y-4 overflow-y-auto px-4 py-4 overscroll-contain" : "space-y-4"}`}>
            {!session?.user_profile.is_hris ? (
              <div>
                <p className="text-xs uppercase tracking-[0.3em] text-slate-500">{/* Perusahaan */}</p>
                <select
                  value={selectedCompany}
                  onChange={(e) => handleCompanyChange(e.target.value)}
                  className="mt-0 w-full border border-slate-200 bg-slate-50 px-4 py-3 text-sm outline-none transition focus:border-slate-400 focus:ring-2 focus:ring-slate-100"
                >
                  {boardData?.companies.map((c, index) => (
                    <option key={c.id || `empty-company-${index}`} value={c.id}>
                      {c.name}
                    </option>
                  ))}
                </select>
              </div>
            ) : null}

            <div>
              <p className="text-xs uppercase tracking-[0.3em] text-slate-500">{/* Modul */}</p>
              <div className="mt-3 overflow-hidden border border-slate-200 bg-slate-50 p-3">
                {selectedCompany ? (
                  renderModuleTree(boardData?.modules ?? [], 0, selectedModule, handleModuleSelect)
                ) : (
                  <p className="px-2 py-6 text-center text-sm text-slate-500">Pilih perusahaan terlebih dahulu.</p>
                )}
              </div>
            </div>
          </div>
        </div>

        <div className="min-w-[800px] flex-1 overflow-y-auto px-3 py-3">
          {error ? <DismissibleAlert tone="error" message={error} onClose={() => setError(null)} className="mb-6 mt-0" /> : null}

          {activeMainView === "admin" && session?.user_profile.is_admin ? (
            <div className="space-y-6">
              <div className="rounded-1xl border border-slate-200 bg-white p-6 shadow-sm">
                <div className="flex flex-wrap items-center justify-between gap-3 border-b border-slate-200">
                  <div>
                    <div className="mt-5 flex flex-wrap gap-2 pb-4">
                      {(["company", "module", "user"] as const).map((tab) => (
                        <button
                          key={tab}
                          type="button"
                          onClick={() => {
                            setActiveAdminTab(tab);
                            setAdminMessage(null);
                            setAdminError(null);
                            resetAdminForms();
                          }}
                          className={`rounded-xl px-4 py-2 text-sm font-medium ${activeAdminTab === tab ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"}`}
                        >
                          {tab === "company" ? "Perusahaan" : tab === "module" ? "Modul" : "Pengguna"}
                        </button>
                      ))}
                    </div>
                  </div>
                  <button
                    type="button"
                    onClick={() => setActiveMainView("workspace")}
                    className="rounded-xl border border-slate-200 px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
                  >
                    Kembali ke Modul
                  </button>
                </div>

                {adminError ? <DismissibleAlert tone="error" message={adminError} onClose={() => setAdminError(null)} /> : null}
                {adminMessage ? <DismissibleAlert tone="success" message={adminMessage} onClose={() => setAdminMessage(null)} /> : null}
                {adminLoading ? <p className="mt-4 text-sm text-slate-500">Memuat data admin...</p> : null}

                {!adminLoading && activeAdminTab === "company" ? (
                  <div className="mt-5 space-y-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <input
                        value={companyFilter}
                        onChange={(e) => setCompanyFilter(e.target.value)}
                        placeholder="Filter perusahaan..."
                        className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm md:w-80"
                      />
                      <button
                        type="button"
                        onClick={() => {
                          resetAdminForms();
                          setAdminFormError(null);
                          setCompanyDialogTab("profile");
                          setShowCompanyDialog(true);
                        }}
                        className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white"
                      >
                        Tambah Perusahaan
                      </button>
                    </div>
                    <div className="overflow-hidden rounded-2xl border border-slate-200">
                      <table className="min-w-full divide-y divide-slate-200 text-sm">
                        <thead className="bg-slate-50">
                          <tr>
                            <th className="w-25 px-3 py-2 text-left font-medium text-slate-600">Kode</th>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Nama</th>
                            <th className="w-25 px-3 py-2 text-left font-medium text-slate-600">Aktif</th>
                            <th className="w-16 px-3 py-2 text-center font-medium text-slate-600">Aksi</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100 bg-white">
                          {companyRows.map((item) => (
                            <tr key={item.id}>
                              <td className="w-25 px-3 py-2">{item.slug}</td>
                              <td className="px-3 py-2">{item.name}</td>
                              <td className="w-25 px-3 py-2">{item.is_active ? "Ya" : "Tidak"}</td>
                              <td className="w-16 px-3 py-2 text-center">
                                <button
                                  type="button"
                                  onClick={async () => {
                                    setCompanyAdminForm(item);
                                    setCompanyDialogTab("profile");
                                    setCompanyModuleAccessForm({ id: "", module_id: "", is_active: true });
                                    setCompanyModuleAccessRows([]);
                                    setAdminFormError(null);
                                    setShowCompanyDialog(true);
                                    try {
                                      await fetchCompanyModuleAccess(item.id);
                                    } catch (err) {
                                      setAdminFormError(
                                        axios.isAxiosError(err)
                                          ? err.response?.data?.error || "Gagal memuat hak akses modul perusahaan."
                                          : "Gagal memuat hak akses modul perusahaan."
                                      );
                                    }
                                  }}
                                  aria-label="Ubah perusahaan"
                                  title="Ubah"
                                  className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-700 transition hover:bg-slate-100"
                                >
                                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                                    <path d="M12 20h9" />
                                    <path d="M16.5 3.5a2.121 2.121 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
                                  </svg>
                                </button>
                              </td>
                            </tr>
                          ))}
                          {companyRows.length === 0 ? (
                            <tr>
                              <td colSpan={4} className="px-3 py-4 text-center text-slate-500">Data tidak ditemukan.</td>
                            </tr>
                          ) : null}
                        </tbody>
                      </table>
                    </div>
                    <div className="flex items-center justify-between rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600">
                      <span>Halaman {companyPage} dari {companyPageTotal}</span>
                      <div className="flex gap-2">
                        <button type="button" disabled={companyPage <= 1} onClick={() => setCompanyPage((prev) => Math.max(1, prev - 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Prev</button>
                        <button type="button" disabled={companyPage >= companyPageTotal} onClick={() => setCompanyPage((prev) => Math.min(companyPageTotal, prev + 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Next</button>
                      </div>
                    </div>
                  </div>
                ) : null}

                {!adminLoading && activeAdminTab === "module" ? (
                  <div className="mt-5 space-y-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <input
                        value={moduleFilter}
                        onChange={(e) => setModuleFilter(e.target.value)}
                        placeholder="Filter modul..."
                        className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm md:w-80"
                      />
                      <button
                        type="button"
                        onClick={() => {
                          resetAdminForms();
                          setAdminFormError(null);
                          setShowModuleDialog(true);
                        }}
                        className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white"
                      >
                        Tambah Modul
                      </button>
                    </div>
                    <div className="overflow-hidden rounded-2xl border border-slate-200">
                      <table className="min-w-full divide-y divide-slate-200 text-sm">
                        <thead className="bg-slate-50">
                          <tr>
                            <th className="w-25 px-3 py-2 text-left font-medium text-slate-600">Kode</th>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Nama</th>
                            <th className="w-16 px-3 py-2 text-left font-medium text-slate-600">Halaman</th>
                            <th className="w-16 px-3 py-2 text-left font-medium text-slate-600">Aktif</th>
                            <th className="w-16 px-3 py-2 text-center font-medium text-slate-600">Aksi</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100 bg-white">
                          {moduleRows.map((item) => (
                            <tr key={item.id}>
                              <td className="w-25 px-3 py-2">{item.code}</td>
                              <td className="px-3 py-2">{item.name}</td>
                              <td className="w-16 px-3 py-2 text-center">{item.is_page ? "Ya" : "Tidak"}</td>
                              <td className="w-16 px-3 py-2 text-center">{item.is_active ? "Ya" : "Tidak"}</td>
                              <td className="w-16 px-3 py-2 text-center">
                                <button
                                  type="button"
                                  onClick={() => { setModuleAdminForm(item); setAdminFormError(null); setShowModuleDialog(true); }}
                                  aria-label="Ubah modul"
                                  title="Ubah"
                                  className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-700 transition hover:bg-slate-100"
                                >
                                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                                    <path d="M12 20h9" />
                                    <path d="M16.5 3.5a2.121 2.121 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
                                  </svg>
                                </button>
                              </td>
                            </tr>
                          ))}
                          {moduleRows.length === 0 ? (
                            <tr>
                              <td colSpan={4} className="px-3 py-4 text-center text-slate-500">Data tidak ditemukan.</td>
                            </tr>
                          ) : null}
                        </tbody>
                      </table>
                    </div>
                    <div className="flex items-center justify-between rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600">
                      <span>Halaman {modulePage} dari {modulePageTotal}</span>
                      <div className="flex gap-2">
                        <button type="button" disabled={modulePage <= 1} onClick={() => setModulePage((prev) => Math.max(1, prev - 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Prev</button>
                        <button type="button" disabled={modulePage >= modulePageTotal} onClick={() => setModulePage((prev) => Math.min(modulePageTotal, prev + 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Next</button>
                      </div>
                    </div>
                  </div>
                ) : null}

                {!adminLoading && activeAdminTab === "user" ? (
                  <div className="mt-5 space-y-4">
                    <div className="flex flex-wrap items-center justify-between gap-3">
                      <input
                        value={userFilter}
                        onChange={(e) => setUserFilter(e.target.value)}
                        placeholder="Filter pengguna..."
                        className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm md:w-80"
                      />
                      <button
                        type="button"
                        onClick={() => {
                          resetAdminForms();
                          setAdminFormError(null);
                          setUserDialogTab("profile");
                          setShowUserDialog(true);
                        }}
                        className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white"
                      >
                        Tambah Pengguna
                      </button>
                    </div>
                    <div className="overflow-hidden rounded-2xl border border-slate-200">
                      <table className="min-w-full divide-y divide-slate-200 text-sm">
                        <thead className="bg-slate-50">
                          <tr>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Perusahaan</th>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Nama Pengguna</th>
                            <th className="w-25 px-3 py-2 text-left font-medium text-slate-600">Aktif</th>
                            <th className="w-16 px-3 py-2 text-center font-medium text-slate-600">Aksi</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100 bg-white">
                          {userRows.map((item) => (
                            <tr key={item.id}>
                              <td className="px-3 py-2">{adminCompanies.find((company) => company.id === item.company_id)?.name || "Multi Perusahaan"}</td>
                              <td className="px-3 py-2">{item.username}</td>
                              <td className="w-25 px-3 py-2">{item.is_active ? "Ya" : "Tidak"}</td>
                              <td className="w-16 px-3 py-2 text-center">
                                <button
                                  type="button"
                                  onClick={async () => {
                                    setUserAdminForm({ ...item, password: "" });
                                    setUserDialogTab("profile");
                                    setUserCompanyAccessRows([]);
                                    setUserPrivilegeAccessRows([]);
                                    setUserCompanyAccessForm({ id: "", company_id: "", is_active: true });
                                    setUserPrivilegeAccessForm({ id: "", user_company_id: "", module_id: "", level: "hide" });
                                    setSelectedUserCompanyAccess("");
                                    setAdminFormError(null);
                                    setShowUserDialog(true);
                                    try {
                                      await fetchUserCompanyAccess(item.id);
                                    } catch (err) {
                                      setAdminFormError(
                                        axios.isAxiosError(err)
                                          ? err.response?.data?.error || "Gagal memuat hak akses pengguna."
                                          : "Gagal memuat hak akses pengguna."
                                      );
                                    }
                                  }}
                                  aria-label="Ubah pengguna"
                                  title="Ubah"
                                  className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-700 transition hover:bg-slate-100"
                                >
                                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                                    <path d="M12 20h9" />
                                    <path d="M16.5 3.5a2.121 2.121 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
                                  </svg>
                                </button>
                              </td>
                            </tr>
                          ))}
                          {userRows.length === 0 ? (
                            <tr>
                              <td colSpan={4} className="px-3 py-4 text-center text-slate-500">Data tidak ditemukan.</td>
                            </tr>
                          ) : null}
                        </tbody>
                      </table>
                    </div>
                    <div className="flex items-center justify-between rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600">
                      <span>Halaman {userPage} dari {userPageTotal}</span>
                      <div className="flex gap-2">
                        <button type="button" disabled={userPage <= 1} onClick={() => setUserPage((prev) => Math.max(1, prev - 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Prev</button>
                        <button type="button" disabled={userPage >= userPageTotal} onClick={() => setUserPage((prev) => Math.min(userPageTotal, prev + 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Next</button>
                      </div>
                    </div>
                  </div>
                ) : null}
              </div>
            </div>
          ) : selectedModuleNode ? (
            <div className="space-y-6">
              <div className="rounded-1xl border border-slate-200 bg-white p-6 shadow-sm">
                <p className="text-xs uppercase tracking-[0.3em] text-slate-500">{selectedModuleNode.path} - {selectedModuleNode.title}</p>
                {isMN01Module ? (
                  <>
                    {selectedModuleNode.level === "hide" ? (
                      <div className="mt-3 rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700">
                        Anda tidak memiliki akses ke modul MN01 (level: hide).
                      </div>
                    ) : null}

                    {selectedModuleNode.level === "view" ? (
                      <div className="mt-3 space-y-4">
                        <div className="flex items-center justify-between rounded-xl border border-sky-200 bg-sky-50 px-4 py-3 text-sm text-sky-800">
                          <span>Modul MN01 sedang berjalan dalam mode baca.</span>
                          <span className="rounded-full bg-sky-100 px-2 py-1 text-xs font-semibold tracking-wide text-sky-700">VIEW</span>
                        </div>
                        <div className="rounded-xl border border-slate-200 p-4">
                          <p className="mb-2 text-sm font-semibold text-slate-900">Contoh Konten MN01 (Read Only)</p>
                          <div className="overflow-hidden rounded-xl border border-slate-200">
                            <table className="min-w-full divide-y divide-slate-200 text-sm">
                              <thead className="bg-slate-50">
                                <tr>
                                  <th className="px-3 py-2 text-left font-medium text-slate-600">KPI</th>
                                  <th className="px-3 py-2 text-left font-medium text-slate-600">Nilai</th>
                                  <th className="px-3 py-2 text-left font-medium text-slate-600">Status</th>
                                </tr>
                              </thead>
                              <tbody className="divide-y divide-slate-100 bg-white">
                                <tr>
                                  <td className="px-3 py-2">Container Health</td>
                                  <td className="px-3 py-2">98.4%</td>
                                  <td className="px-3 py-2">Normal</td>
                                </tr>
                                <tr>
                                  <td className="px-3 py-2">CPU Alert</td>
                                  <td className="px-3 py-2">2 node</td>
                                  <td className="px-3 py-2">Perlu review</td>
                                </tr>
                              </tbody>
                            </table>
                          </div>
                        </div>
                      </div>
                    ) : null}

                    {selectedModuleNode.level === "post" ? (
                      <div className="mt-3 space-y-4">
                        <div className="flex items-center justify-between rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">
                          <span>Modul MN01 aktif dalam mode input data.</span>
                          <span className="rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold tracking-wide text-emerald-700">POST</span>
                        </div>
                        <form className="grid gap-3 rounded-xl border border-slate-200 p-4" onSubmit={handleMN01Submit}>
                          <label className="grid gap-1 text-sm text-slate-700">
                            <span>Judul Catatan Monitoring</span>
                            <input
                              value={mn01PostForm.title}
                              onChange={(e) => setMn01PostForm((prev) => ({ ...prev, title: e.target.value }))}
                              className="rounded-xl border border-slate-200 px-3 py-2"
                            />
                          </label>
                          <label className="grid gap-1 text-sm text-slate-700">
                            <span>Catatan</span>
                            <textarea
                              value={mn01PostForm.note}
                              onChange={(e) => setMn01PostForm((prev) => ({ ...prev, note: e.target.value }))}
                              className="min-h-24 rounded-xl border border-slate-200 px-3 py-2"
                            />
                          </label>
                          <div className="flex gap-2">
                            <button type="submit" className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white">Simpan Catatan</button>
                            <button type="button" onClick={() => setMn01PostForm({ title: "", note: "" })} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">Reset</button>
                          </div>
                        </form>

                        <div className="overflow-hidden rounded-xl border border-slate-200">
                          <table className="min-w-full divide-y divide-slate-200 text-sm">
                            <thead className="bg-slate-50">
                              <tr>
                                <th className="px-3 py-2 text-left font-medium text-slate-600">Waktu</th>
                                <th className="px-3 py-2 text-left font-medium text-slate-600">Judul</th>
                                <th className="px-3 py-2 text-left font-medium text-slate-600">Catatan</th>
                              </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-100 bg-white">
                              {mn01PostRows.length === 0 ? (
                                <tr>
                                  <td colSpan={3} className="px-3 py-4 text-center text-slate-500">Belum ada catatan monitoring.</td>
                                </tr>
                              ) : (
                                mn01PostRows.map((row) => (
                                  <tr key={row.id}>
                                    <td className="px-3 py-2">{formatDateTime(row.createdAt)}</td>
                                    <td className="px-3 py-2">{row.title}</td>
                                    <td className="px-3 py-2">{row.note}</td>
                                  </tr>
                                ))
                              )}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    ) : null}
                  </>
                ) : (
                  <div className="mt-2 rounded-1xl bg-slate-50 p-3">
                    <p className="text-sm text-slate-700">
                      Modul ini menampilkan konten utama. Detail implementasi modul akan ditampilkan di sini sesuai dengan struktur yang dipilih.
                    </p>
                  </div>
                )}
              </div>
            </div>
          ) : (
            <div className="rounded-1xl border border-slate-200 bg-white p-8 text-center shadow-sm">
              <p className="text-sm text-slate-600">Pilih modul di sebelah kiri untuk melihat detail.</p>
            </div>
          )}
        </div>
      </div>

      {showCompanyDialog ? (
        <div className="fixed inset-0 z-40 bg-black/40 backdrop-blur-sm">
          <div className="h-full w-full overflow-y-auto bg-white p-5 md:p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">{companyAdminForm.id ? "Ubah Perusahaan" : "Tambah Perusahaan"}</h3>
              <button
                type="button"
                onClick={() => {
                  setShowCompanyDialog(false);
                  resetAdminForms();
                }}
                aria-label="Tutup popup"
                title="Tutup"
                className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                  <path d="M18 6 6 18" />
                  <path d="m6 6 12 12" />
                </svg>
              </button>
            </div>
            <div className="mt-4 flex flex-wrap gap-2 border-b border-slate-200 pb-3">
              <button
                type="button"
                onClick={() => setCompanyDialogTab("profile")}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${companyDialogTab === "profile" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"}`}
              >
                Profil Perusahaan
              </button>
              <button
                type="button"
                disabled={!companyAdminForm.id}
                onClick={async () => {
                  setCompanyDialogTab("access");
                  if (companyAdminForm.id) {
                    try {
                      await fetchCompanyModuleAccess(companyAdminForm.id);
                    } catch (err) {
                      setAdminFormError(
                        axios.isAxiosError(err)
                          ? err.response?.data?.error || "Gagal memuat hak akses perusahaan."
                          : "Gagal memuat hak akses perusahaan."
                      );
                    }
                  }
                }}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${companyDialogTab === "access" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"} disabled:cursor-not-allowed disabled:opacity-50`}
              >
                Hak Akses
              </button>
            </div>

            {adminFormError ? <DismissibleAlert tone="error" message={adminFormError} onClose={() => setAdminFormError(null)} className="mt-4" /> : null}

            {companyDialogTab === "profile" ? (
              <form className="mt-4 grid gap-4" onSubmit={saveAdminCompany}>
                <div className="grid gap-3">
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Kode</span>
                    <input value={companyAdminForm.slug} onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, slug: e.target.value }))} disabled={Boolean(companyAdminForm.id)} className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Nama</span>
                    <input value={companyAdminForm.name} onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, name: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Mata Uang</span>
                    <input value={companyAdminForm.valuta} onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, valuta: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>NPWP</span>
                    <input value={companyAdminForm.vat_id} onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, vat_id: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>No. Registrasi</span>
                    <input value={companyAdminForm.reg_no} onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, reg_no: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Tautan HRIS</span>
                    <input value={companyAdminForm.hris_link} onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, hris_link: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-start">
                    <span className="md:pt-2">Alamat</span>
                    <textarea value={companyAdminForm.address} onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, address: e.target.value }))} className="min-h-24 rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Status</span>
                    <span className="flex items-center gap-3 py-1">
                      <input
                        type="checkbox"
                        checked={companyAdminForm.is_active}
                        onChange={(e) => setCompanyAdminForm((prev) => ({ ...prev, is_active: e.target.checked }))}
                        className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300"
                      />
                      Aktif
                    </span>
                  </label>
                </div>
                <div className="mt-4 flex gap-2 pt-1">
                  <button type="submit" disabled={adminBusy} className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white">{adminBusy ? "Menyimpan..." : companyAdminForm.id ? "Ubah Perusahaan" : "Tambah Perusahaan"}</button>
                  <button type="button" onClick={resetAdminForms} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">Reset</button>
                </div>
              </form>
            ) : (
              <div className="mt-4 space-y-4">
                {companyAdminForm.id ? (
                  <>
                    <form className="grid gap-3 rounded-xl border border-slate-200 p-3" onSubmit={saveCompanyModuleAccess}>
                      {companyModuleAccessForm.id ? (
                        <div className="flex items-center justify-between gap-3 rounded-xl border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-800">
                          <span>Sedang mengubah hak akses perusahaan. Pengisian Modul dikunci agar relasi tetap konsisten.</span>
                          <span className="rounded-full bg-sky-100 px-2 py-1 text-xs font-semibold tracking-wide text-sky-700">Mode Ubah</span>
                        </div>
                      ) : null}
                      {companyModuleInputOptions.length === 0 && !companyModuleAccessForm.id ? (
                        <div className="flex items-center justify-between gap-3 rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
                          <span>Semua modul halaman sudah masuk ke hak akses perusahaan ini.</span>
                          <span className="rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold tracking-wide text-emerald-700">Semua Terdaftar</span>
                        </div>
                      ) : null}
                      <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                        <span>Modul</span>
                        <select
                          value={companyModuleAccessForm.module_id ?? ""}
                          onChange={(e) => setCompanyModuleAccessForm((prev) => ({ ...prev, module_id: e.target.value }))}
                          disabled={Boolean(companyModuleAccessForm.id)}
                          className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500"
                        >
                          <option value="">Pilih modul</option>
                          {companyModuleInputOptions.length === 0 ? <option value="" disabled>Semua modul halaman sudah terdaftar</option> : null}
                          {companyModuleInputOptions.map((item) => (
                            <option key={item.id} value={item.id}>{item.code} - {item.name}</option>
                          ))}
                        </select>
                      </label>
                      <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                        <span>Status</span>
                        <span className="flex items-center gap-3 py-1">
                          <input
                            type="checkbox"
                            checked={normalizeBool(companyModuleAccessForm.is_active)}
                            onChange={(e) => setCompanyModuleAccessForm((prev) => ({ ...prev, is_active: e.target.checked }))}
                            className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300"
                          />
                          Aktif
                        </span>
                      </label>
                      <div className="flex gap-2">
                        <button type="submit" disabled={adminBusy || (!companyModuleAccessForm.id && companyModuleInputOptions.length === 0)} className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50">{companyModuleAccessForm.id ? "Ubah Hak Akses" : "Tambah Hak Akses"}</button>
                        <button type="button" onClick={() => setCompanyModuleAccessForm({ id: "", module_id: "", is_active: true })} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">Reset</button>
                      </div>
                    </form>

                    <div className="overflow-hidden rounded-2xl border border-slate-200">
                      <table className="min-w-full divide-y divide-slate-200 text-sm">
                        <thead className="bg-slate-50">
                          <tr>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Kode Modul</th>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Nama Modul</th>
                            <th className="w-20 px-3 py-2 text-left font-medium text-slate-600">Aktif</th>
                            <th className="w-16 px-3 py-2 text-center font-medium text-slate-600">Aksi</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100 bg-white">
                          {companyModuleAccessRows.length === 0 ? (
                            <tr>
                              <td colSpan={4} className="px-3 py-4 text-center text-slate-500">Belum ada hak akses modul untuk perusahaan ini.</td>
                            </tr>
                          ) : (
                            companyModuleAccessRows.map((item) => (
                              <tr key={item.id}>
                                <td className="px-3 py-2">{item.module_code}</td>
                                <td className="px-3 py-2">{item.module_name}</td>
                                <td className="px-3 py-2">{item.is_active ? "Ya" : "Tidak"}</td>
                                <td className="px-3 py-2 text-center">
                                  <button
                                    type="button"
                                    onClick={() => setCompanyModuleAccessForm({ id: item.id ?? "", module_id: item.module_id ?? "", is_active: normalizeBool(item.is_active) })}
                                    aria-label="Ubah hak akses perusahaan"
                                    className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-700 transition hover:bg-slate-100"
                                  >
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                                      <path d="M12 20h9" />
                                      <path d="M16.5 3.5a2.121 2.121 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
                                    </svg>
                                  </button>
                                </td>
                              </tr>
                            ))
                          )}
                        </tbody>
                      </table>
                    </div>
                  </>
                ) : (
                  <p className="text-sm text-slate-600">Simpan data profil perusahaan terlebih dahulu, lalu atur Hak Akses pada tab ini.</p>
                )}
              </div>
            )}
          </div>
        </div>
      ) : null}

      {showModuleDialog ? (
        <div className="fixed inset-0 z-40 bg-black/40 backdrop-blur-sm">
          <div className="h-full w-full overflow-y-auto bg-white p-5 md:p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">{moduleAdminForm.id ? "Ubah Modul" : "Tambah Modul"}</h3>
              <button
                type="button"
                onClick={() => {
                  setShowModuleDialog(false);
                  resetAdminForms();
                }}
                aria-label="Tutup popup"
                title="Tutup"
                className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                  <path d="M18 6 6 18" />
                  <path d="m6 6 12 12" />
                </svg>
              </button>
            </div>
            <form className="mt-4 grid gap-4" onSubmit={saveAdminModule}>
              {adminFormError ? <DismissibleAlert tone="error" message={adminFormError} onClose={() => setAdminFormError(null)} className="mt-0" /> : null}
              <div className="grid gap-3">
                <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                  <span>Induk Modul</span>
                  <select value={moduleAdminForm.parent_id} onChange={(e) => setModuleAdminForm((prev) => ({ ...prev, parent_id: e.target.value }))} disabled={Boolean(moduleAdminForm.id)} className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500">
                    <option value="">Tanpa Induk</option>
                    {adminModules.filter((item) => item.id !== moduleAdminForm.id).map((item) => (
                      <option key={item.id} value={item.id}>{item.name}</option>
                    ))}
                  </select>
                </label>
                <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                  <span>Kode</span>
                  <input value={moduleAdminForm.code} onChange={(e) => setModuleAdminForm((prev) => ({ ...prev, code: e.target.value }))} disabled={Boolean(moduleAdminForm.id)} className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500" />
                </label>
                <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                  <span>Nama</span>
                  <input value={moduleAdminForm.name} onChange={(e) => setModuleAdminForm((prev) => ({ ...prev, name: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                </label>
                <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                  <span>Lokasi</span>
                  <input value={moduleAdminForm.path} onChange={(e) => setModuleAdminForm((prev) => ({ ...prev, path: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                </label>
                <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                  <span>Halaman</span>
                  <span className="flex items-center gap-3 py-1">
                    <input
                      type="checkbox"
                      checked={moduleAdminForm.is_page}
                      onChange={(e) => setModuleAdminForm((prev) => ({ ...prev, is_page: e.target.checked }))}
                      disabled={Boolean(moduleAdminForm.id)}
                      className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300"
                    />
                    Ya
                  </span>
                </label>
                <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                  <span>Status</span>
                  <span className="flex items-center gap-3 py-1">
                    <input
                      type="checkbox"
                      checked={moduleAdminForm.is_active}
                      onChange={(e) => setModuleAdminForm((prev) => ({ ...prev, is_active: e.target.checked }))}
                      className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300"
                    />
                    Aktif
                  </span>
                </label>
              </div>
              <div className="mt-4 flex gap-2 pt-1">
                <button type="submit" disabled={adminBusy} className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white">{adminBusy ? "Menyimpan..." : moduleAdminForm.id ? "Ubah Modul" : "Tambah Modul"}</button>
                <button type="button" onClick={resetAdminForms} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">Reset</button>
              </div>
            </form>
          </div>
        </div>
      ) : null}

      {showUserDialog ? (
        <div className="fixed inset-0 z-40 bg-black/40 backdrop-blur-sm">
          <div className="h-full w-full overflow-y-auto bg-white p-5 md:p-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-slate-900">{userAdminForm.id ? "Ubah Pengguna" : "Tambah Pengguna"}</h3>
              <button
                type="button"
                onClick={() => {
                  setShowUserDialog(false);
                  resetAdminForms();
                }}
                aria-label="Tutup popup"
                title="Tutup"
                className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-600 hover:bg-slate-50"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                  <path d="M18 6 6 18" />
                  <path d="m6 6 12 12" />
                </svg>
              </button>
            </div>
            <div className="mt-4 flex flex-wrap gap-2 border-b border-slate-200 pb-3">
              <button
                type="button"
                onClick={() => setUserDialogTab("profile")}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${userDialogTab === "profile" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"}`}
              >
                Profil Pengguna
              </button>
              <button
                type="button"
                disabled={!userAdminForm.id || isProfileCompanyLocked}
                onClick={async () => {
                  setUserDialogTab("companyAccess");
                  if (userAdminForm.id) {
                    try {
                      await fetchUserCompanyAccess(userAdminForm.id);
                    } catch (err) {
                      setAdminFormError(
                        axios.isAxiosError(err)
                          ? err.response?.data?.error || "Gagal memuat hak akses pengguna."
                          : "Gagal memuat hak akses pengguna."
                      );
                    }
                  }
                }}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${userDialogTab === "companyAccess" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"} disabled:cursor-not-allowed disabled:opacity-50`}
              >
                Akses Perusahaan
              </button>
              <button
                type="button"
                disabled={!userAdminForm.id}
                onClick={async () => {
                  setUserDialogTab("moduleAccess");
                  if (userAdminForm.id) {
                    try {
                      await fetchUserCompanyAccess(userAdminForm.id);
                    } catch (err) {
                      setAdminFormError(
                        axios.isAxiosError(err)
                          ? err.response?.data?.error || "Gagal memuat hak akses pengguna."
                          : "Gagal memuat hak akses pengguna."
                      );
                    }
                  }
                }}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${userDialogTab === "moduleAccess" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"} disabled:cursor-not-allowed disabled:opacity-50`}
              >
                Akses Modul
              </button>
            </div>

            {adminFormError ? <DismissibleAlert tone="error" message={adminFormError} onClose={() => setAdminFormError(null)} className="mt-4" /> : null}

            {userDialogTab === "profile" ? (
              <form className="mt-4 grid gap-4" onSubmit={saveAdminUser}>
                <div className="grid gap-3">
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Perusahaan</span>
                    <select value={userAdminForm.company_id} onChange={(e) => setUserAdminForm((prev) => ({ ...prev, company_id: e.target.value }))} disabled={Boolean(userAdminForm.id)} className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500">
                      <option value="">Multi Perusahaan</option>
                      {adminCompanies.map((item) => (
                        <option key={item.id} value={item.id}>{item.name}</option>
                      ))}
                    </select>
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Nama Pengguna</span>
                    <input value={userAdminForm.username} onChange={(e) => setUserAdminForm((prev) => ({ ...prev, username: e.target.value }))} disabled={Boolean(userAdminForm.id)} className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Email</span>
                    <input value={userAdminForm.email} onChange={(e) => setUserAdminForm((prev) => ({ ...prev, email: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Nama Lengkap</span>
                    <input value={userAdminForm.fullname} onChange={(e) => setUserAdminForm((prev) => ({ ...prev, fullname: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>No. Ponsel</span>
                    <input value={userAdminForm.phone} onChange={(e) => setUserAdminForm((prev) => ({ ...prev, phone: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Jabatan</span>
                    <input value={userAdminForm.role} onChange={(e) => setUserAdminForm((prev) => ({ ...prev, role: e.target.value }))} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Kata Sandi</span>
                    <input type="password" value={userAdminForm.password || ""} onChange={(e) => setUserAdminForm((prev) => ({ ...prev, password: e.target.value }))} placeholder={userAdminForm.id ? "Kata Sandi baru (opsional)" : "Kata Sandi"} className="rounded-xl border border-slate-200 px-3 py-2" />
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Admin</span>
                    <span className="flex items-center gap-3 py-1">
                      <input
                        type="checkbox"
                        checked={userAdminForm.is_admin}
                        onChange={(e) => setUserAdminForm((prev) => ({ ...prev, is_admin: e.target.checked }))}
                        className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300"
                      />
                      Ya
                    </span>
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>HRIS</span>
                    <span className="flex items-center gap-3 py-1">
                      <input
                        type="checkbox"
                        checked={userAdminForm.is_hris}
                        onChange={(e) => setUserAdminForm((prev) => ({ ...prev, is_hris: e.target.checked }))}
                        className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300"
                      />
                      Ya
                    </span>
                  </label>
                  <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                    <span>Status</span>
                    <span className="flex items-center gap-3 py-1">
                      <input
                        type="checkbox"
                        checked={userAdminForm.is_active}
                        onChange={(e) => setUserAdminForm((prev) => ({ ...prev, is_active: e.target.checked }))}
                        className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300"
                      />
                      Aktif
                    </span>
                  </label>
                </div>
                <div className="mt-4 flex gap-2 pt-1">
                  <button type="submit" disabled={adminBusy} className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white">{adminBusy ? "Menyimpan..." : userAdminForm.id ? "Ubah Pengguna" : "Tambah Pengguna"}</button>
                  <button type="button" onClick={resetAdminForms} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">Reset</button>
                </div>
              </form>
            ) : (
              <div className="mt-4 space-y-4">
                {userAdminForm.id ? (
                  <>
                    {userDialogTab === "companyAccess" ? (
                      <>
                        {isProfileCompanyLocked ? (
                          <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-700">
                            Akses Perusahaan dinonaktifkan karena profil pengguna sudah memiliki perusahaan tetap.
                          </p>
                        ) : null}
                    <form className="grid gap-3 rounded-xl border border-slate-200 p-3" onSubmit={saveUserCompanyAccess}>
                      {userCompanyAccessForm.id ? (
                        <div className="flex items-center justify-between gap-3 rounded-xl border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-800">
                          <span>Sedang mengubah akses perusahaan pengguna. Pengisian Perusahaan dikunci agar relasi tetap konsisten.</span>
                          <span className="rounded-full bg-sky-100 px-2 py-1 text-xs font-semibold tracking-wide text-sky-700">Mode Ubah</span>
                        </div>
                      ) : null}
                      {userCompanyInputOptions.length === 0 && !userCompanyAccessForm.id ? (
                        <div className="flex items-center justify-between gap-3 rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
                          <span>Semua perusahaan sudah dimasukkan pada akses perusahaan pengguna ini.</span>
                          <span className="rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold tracking-wide text-emerald-700">Semua Terdaftar</span>
                        </div>
                      ) : null}
                      <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                        <span>Perusahaan</span>
                        <select
                          value={userCompanyAccessForm.company_id ?? ""}
                          onChange={(e) => setUserCompanyAccessForm((prev) => ({ ...prev, company_id: e.target.value }))}
                          disabled={isProfileCompanyLocked || Boolean(userCompanyAccessForm.id)}
                          className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500"
                        >
                          <option value="">Pilih perusahaan</option>
                          {userCompanyInputOptions.length === 0 ? <option value="" disabled>Semua perusahaan sudah terdaftar</option> : null}
                          {userCompanyInputOptions.map((item) => (
                            <option key={item.id} value={item.id}>{item.name}</option>
                          ))}
                        </select>
                      </label>
                      <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                        <span>Status</span>
                        <span className="flex items-center gap-3 py-1">
                          <input
                            type="checkbox"
                            checked={normalizeBool(userCompanyAccessForm.is_active)}
                            onChange={(e) => setUserCompanyAccessForm((prev) => ({ ...prev, is_active: e.target.checked }))}
                            disabled={isProfileCompanyLocked}
                            className="h-5 w-5 rounded border-slate-300 text-slate-900 focus:ring-2 focus:ring-slate-300 disabled:opacity-50"
                          />
                          Aktif
                        </span>
                      </label>
                      <div className="flex gap-2">
                        <button type="submit" disabled={adminBusy || isProfileCompanyLocked || (!userCompanyAccessForm.id && userCompanyInputOptions.length === 0)} className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50">Akses Perusahaan</button>
                        <button type="button" disabled={isProfileCompanyLocked} onClick={() => setUserCompanyAccessForm({ id: "", company_id: "", is_active: true })} className="rounded-xl border border-slate-200 px-4 py-2 text-sm disabled:opacity-50">Reset</button>
                      </div>
                    </form>

                    <div className="overflow-hidden rounded-2xl border border-slate-200">
                      <table className="min-w-full divide-y divide-slate-200 text-sm">
                        <thead className="bg-slate-50">
                          <tr>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Perusahaan</th>
                            <th className="w-20 px-3 py-2 text-left font-medium text-slate-600">Aktif</th>
                            <th className="w-16 px-3 py-2 text-center font-medium text-slate-600">Aksi</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100 bg-white">
                          {userCompanyAccessRows.length === 0 ? (
                            <tr>
                              <td colSpan={3} className="px-3 py-4 text-center text-slate-500">Belum ada akses perusahaan pengguna.</td>
                            </tr>
                          ) : (
                            userCompanyAccessRows.map((item) => (
                              <tr key={item.id}>
                                <td className="px-3 py-2">{item.company_name}</td>
                                <td className="px-3 py-2">{item.is_active ? "Ya" : "Tidak"}</td>
                                <td className="px-3 py-2 text-center">
                                  <button
                                    type="button"
                                    disabled={isProfileCompanyLocked}
                                    onClick={() => {
                                      setUserCompanyAccessForm({ id: item.id ?? "", company_id: item.company_id ?? "", is_active: normalizeBool(item.is_active) });
                                      setSelectedUserCompanyAccess(item.id);
                                    }}
                                    aria-label="Ubah akses perusahaan pengguna"
                                    className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-700 transition hover:bg-slate-100 disabled:opacity-50"
                                  >
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                                      <path d="M12 20h9" />
                                      <path d="M16.5 3.5a2.121 2.121 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
                                    </svg>
                                  </button>
                                </td>
                              </tr>
                            ))
                          )}
                        </tbody>
                      </table>
                    </div>
                      </>
                    ) : null}

                    {userDialogTab === "moduleAccess" ? (
                      <>
                    {isProfileCompanyLocked && !hasProfileCompanyMapping ? (
                      <p className="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-700">
                        Pemetaan perusahaan pengguna untuk perusahaan pada profil pengguna belum tersedia. Tambahkan relasi di data perusahaan pengguna terlebih dahulu.
                      </p>
                    ) : null}
                    <form className="grid gap-3 rounded-xl border border-slate-200 p-3" onSubmit={saveUserPrivilegeAccess}>
                      {userPrivilegeAccessForm.id ? (
                        <div className="flex items-center justify-between gap-3 rounded-xl border border-sky-200 bg-sky-50 px-3 py-2 text-sm text-sky-800">
                          <span>Sedang mengubah akses modul pengguna. Pengisian Perusahaan dan Modul dikunci agar relasi tetap konsisten.</span>
                          <span className="rounded-full bg-sky-100 px-2 py-1 text-xs font-semibold tracking-wide text-sky-700">Mode Ubah</span>
                        </div>
                      ) : null}
                      {userPrivilegeModuleInputOptions.length === 0 && !userPrivilegeAccessForm.id ? (
                        <div className="flex items-center justify-between gap-3 rounded-xl border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-800">
                          <span>Semua modul aktif perusahaan ini sudah ada di daftar akses modul.</span>
                          <span className="rounded-full bg-emerald-100 px-2 py-1 text-xs font-semibold tracking-wide text-emerald-700">Semua Terdaftar</span>
                        </div>
                      ) : null}
                      <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                        <span>Perusahaan</span>
                        <select
                          value={isProfileCompanyLocked ? selectedUserCompanyAccess : (userPrivilegeAccessForm.user_company_id || selectedUserCompanyAccess)}
                          onChange={(e) => {
                            if (isProfileCompanyLocked) return;
                            const value = e.target.value;
                            setUserPrivilegeAccessForm((prev) => ({ ...prev, user_company_id: value }));
                            setSelectedUserCompanyAccess(value);
                          }}
                          disabled={isProfileCompanyLocked || Boolean(userPrivilegeAccessForm.id)}
                          className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500"
                        >
                          <option value="">Pilih perusahaan</option>
                          {userCompanyAccessRows.map((item) => (
                            <option key={item.id} value={item.id}>{item.company_name}</option>
                          ))}
                        </select>
                      </label>
                      <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                        <span>Modul</span>
                        <select
                          value={userPrivilegeAccessForm.module_id}
                          onChange={(e) => setUserPrivilegeAccessForm((prev) => ({ ...prev, module_id: e.target.value }))}
                          disabled={Boolean(userPrivilegeAccessForm.id)}
                          className="rounded-xl border border-slate-200 px-3 py-2 disabled:bg-slate-100 disabled:text-slate-500"
                        >
                          <option value="">Pilih modul</option>
                          {userPrivilegeModuleInputOptions.length === 0 ? <option value="" disabled>Semua modul aktif sudah terdaftar</option> : null}
                          {userPrivilegeModuleInputOptions.map((item) => (
                            <option key={item.id} value={item.id}>{item.code} - {item.name}</option>
                          ))}
                        </select>
                      </label>
                      <label className="grid gap-1 text-sm text-slate-700 md:grid-cols-[170px_minmax(0,1fr)] md:items-center">
                        <span>Tingkatan</span>
                        <select
                          value={userPrivilegeAccessForm.level}
                          onChange={(e) => setUserPrivilegeAccessForm((prev) => ({ ...prev, level: e.target.value as "hide" | "view" | "book" | "post" }))}
                          className="rounded-xl border border-slate-200 px-3 py-2"
                        >
                          <option value="hide">Sembunyi</option>
                          <option value="view">Lihat</option>
                          <option value="book">Pesan</option>
                          <option value="post">Proses</option>
                        </select>
                      </label>
                      <div className="flex gap-2">
                        <button type="submit" disabled={adminBusy || !hasProfileCompanyMapping || (!userPrivilegeAccessForm.id && userPrivilegeModuleInputOptions.length === 0)} className="rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white disabled:opacity-50">Akses Modul</button>
                        <button type="button" onClick={() => setUserPrivilegeAccessForm({ id: "", user_company_id: selectedUserCompanyAccess, module_id: "", level: "hide" })} className="rounded-xl border border-slate-200 px-4 py-2 text-sm">Reset</button>
                      </div>
                    </form>

                    <div className="overflow-hidden rounded-2xl border border-slate-200">
                      <table className="min-w-full divide-y divide-slate-200 text-sm">
                        <thead className="bg-slate-50">
                          <tr>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Perusahaan</th>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Kode</th>
                            <th className="px-3 py-2 text-left font-medium text-slate-600">Nama</th>
                            <th className="w-20 px-3 py-2 text-left font-medium text-slate-600">Tingkatan</th>
                            <th className="w-16 px-3 py-2 text-center font-medium text-slate-600">Aksi</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100 bg-white">
                          {userPrivilegeAccessRows.length === 0 ? (
                            <tr>
                              <td colSpan={5} className="px-3 py-4 text-center text-slate-500">Belum ada akses modul pengguna untuk perusahaan terpilih.</td>
                            </tr>
                          ) : (
                            userPrivilegeAccessRows.map((item) => (
                              <tr key={item.id}>
                                <td className="px-3 py-2">{selectedPrivilegeCompany ? `${selectedPrivilegeCompany.slug} - ${selectedPrivilegeCompany.name}` : "-"}</td>
                                <td className="px-3 py-2">{item.module_code}</td>
                                <td className="px-3 py-2">{item.module_name}</td>
                                <td className="px-3 py-2">{userPrivilegeAccessFormLabels[item.level] || item.level}</td>
                                <td className="px-3 py-2 text-center">
                                  <button
                                    type="button"
                                    onClick={() => setUserPrivilegeAccessForm({ id: item.id, user_company_id: item.user_company_id, module_id: item.module_id, level: item.level })}
                                    aria-label="Ubah akses modul pengguna"
                                    className="inline-flex h-9 w-9 items-center justify-center rounded-lg border border-slate-200 text-slate-700 transition hover:bg-slate-100"
                                  >
                                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                                      <path d="M12 20h9" />
                                      <path d="M16.5 3.5a2.121 2.121 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
                                    </svg>
                                  </button>
                                </td>
                              </tr>
                            ))
                          )}
                        </tbody>
                      </table>
                    </div>
                      </>
                    ) : null}
                  </>
                ) : (
                  <p className="text-sm text-slate-600">Simpan data profil pengguna terlebih dahulu, lalu atur Hak Akses pada tab ini.</p>
                )}
              </div>
            )}
          </div>
        </div>
      ) : null}

      {showProfileModal ? (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4 backdrop-blur-sm">
          <div className="w-full max-w-3xl rounded-3xl border border-slate-200 bg-white p-6 shadow-2xl">
            <div className="flex items-center justify-between">
              <h2 className="text-2xl font-semibold text-slate-900">Profil</h2>
              <button
                type="button"
                onClick={() => setShowProfileModal(false)}
                aria-label="Tutup popup"
                title="Tutup"
                className="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-slate-200 text-slate-600 transition hover:bg-slate-50"
              >
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="h-4 w-4" aria-hidden="true">
                  <path d="M18 6 6 18" />
                  <path d="m6 6 12 12" />
                </svg>
              </button>
            </div>

            <div className="mt-5 flex flex-wrap gap-2 border-b border-slate-200 pb-4">
              <button
                type="button"
                onClick={() => setActiveProfileTab("info")}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${activeProfileTab === "info" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"}`}
              >
                Info
              </button>
              <button
                type="button"
                onClick={() => setActiveProfileTab("profile")}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${activeProfileTab === "profile" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"}`}
              >
                Profil
              </button>
              {!session?.user_profile.is_hris ? (
                <button
                  type="button"
                  onClick={() => setActiveProfileTab("password")}
                  className={`rounded-xl px-4 py-2 text-sm font-medium ${activeProfileTab === "password" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"}`}
                >
                  Kata Sandi
                </button>
              ) : null}
              <button
                type="button"
                onClick={() => setActiveProfileTab("history")}
                className={`rounded-xl px-4 py-2 text-sm font-medium ${activeProfileTab === "history" ? "bg-slate-900 text-white" : "bg-slate-100 text-slate-700"}`}
              >
                Riwayat
              </button>
            </div>

            {profileError ? <DismissibleAlert tone="error" message={profileError} onClose={() => setProfileError(null)} /> : null}
            {profileMessage ? <DismissibleAlert tone="success" message={profileMessage} onClose={() => setProfileMessage(null)} /> : null}

            <div className="mt-5">
              {profileLoading ? <p className="text-sm text-slate-500">Memuat data...</p> : null}

              {!profileLoading && activeProfileTab === "info" ? (
                <div className="grid gap-4 md:grid-cols-2">
                  <div className="rounded-2xl bg-slate-50 p-4 md:col-span-2">
                    <p className="mt-2 text-sm font-semibold text-slate-900">{profileData?.fullname || session?.user_profile.fullname}</p>
                    <p className="mt-2 text-xs text-slate-600">Nama Pengguna: {profileData?.username || session?.user_profile.username}</p>
                    <p className="mt-2 text-xs text-slate-600">Email: {profileData?.email || session?.user_profile.email}</p>
                    <p className="mt-2 text-xs text-slate-600">Ponsel: {profileData?.phone || "-"}</p>
                    <p className="mt-2 text-xs font-semibold text-slate-900">{company.slug} - {company.name}</p>
                    <p className="mt-2 break-all text-xs font-mono text-slate-700">{session?.token}</p>
                    <p className="mt-2 text-xs text-slate-500">
                      {session ? formatDateTime(session.expires_at) : "-"}
                    </p>
                  </div>
                </div>
              ) : null}

              {!profileLoading && activeProfileTab === "profile" ? (
                <form className="grid gap-4" onSubmit={handleProfileUpdate}>
                  <label className="grid gap-1 text-sm text-slate-700">
                    Nama Lengkap
                    <input
                      type="text"
                      value={profileForm.fullname}
                      onChange={(e) => setProfileForm((prev) => ({ ...prev, fullname: e.target.value }))}
                      className="rounded-xl border border-slate-200 px-3 py-2 outline-none focus:border-slate-400"
                    />
                  </label>

                  <label className="grid gap-1 text-sm text-slate-700">
                    Email
                    <input
                      type="email"
                      value={profileForm.email}
                      onChange={(e) => setProfileForm((prev) => ({ ...prev, email: e.target.value }))}
                      className="rounded-xl border border-slate-200 px-3 py-2 outline-none focus:border-slate-400"
                    />
                  </label>

                  <label className="grid gap-1 text-sm text-slate-700">
                    No. Ponsel
                    <input
                      type="text"
                      value={profileForm.phone}
                      onChange={(e) => setProfileForm((prev) => ({ ...prev, phone: e.target.value }))}
                      className="rounded-xl border border-slate-200 px-3 py-2 outline-none focus:border-slate-400"
                    />
                  </label>

                  <button
                    type="submit"
                    disabled={profileBusy}
                    className="mt-2 rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800 disabled:opacity-50"
                  >
                    {profileBusy ? "Menyimpan..." : "Simpan Profil"}
                  </button>
                </form>
              ) : null}

              {!profileLoading && activeProfileTab === "password" && !session?.user_profile.is_hris ? (
                <form className="grid gap-4" onSubmit={handlePasswordUpdate}>
                  <label className="grid gap-1 text-sm text-slate-700">
                    Kata Sandi Saat Ini
                    <input
                      type="password"
                      value={passwordForm.current_password}
                      onChange={(e) => setPasswordForm((prev) => ({ ...prev, current_password: e.target.value }))}
                      className="rounded-xl border border-slate-200 px-3 py-2 outline-none focus:border-slate-400"
                    />
                  </label>

                  <label className="grid gap-1 text-sm text-slate-700">
                    Kata Sandi Baru
                    <input
                      type="password"
                      value={passwordForm.new_password}
                      onChange={(e) => setPasswordForm((prev) => ({ ...prev, new_password: e.target.value }))}
                      className="rounded-xl border border-slate-200 px-3 py-2 outline-none focus:border-slate-400"
                    />
                  </label>

                  <label className="grid gap-1 text-sm text-slate-700">
                    Konfirmasi Kata Sandi Baru
                    <input
                      type="password"
                      value={passwordForm.confirm_password}
                      onChange={(e) => setPasswordForm((prev) => ({ ...prev, confirm_password: e.target.value }))}
                      className="rounded-xl border border-slate-200 px-3 py-2 outline-none focus:border-slate-400"
                    />
                  </label>

                  <button
                    type="submit"
                    disabled={profileBusy}
                    className="mt-2 rounded-xl bg-slate-900 px-4 py-2 text-sm font-semibold text-white hover:bg-slate-800 disabled:opacity-50"
                  >
                    {profileBusy ? "Memperbarui..." : "Ubah Kata Sandi"}
                  </button>
                </form>
              ) : null}

              {!profileLoading && activeProfileTab === "history" ? (
                <div className="space-y-4">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <input
                      value={historyFilter}
                      onChange={(e) => setHistoryFilter(e.target.value)}
                      placeholder="Filter riwayat..."
                      className="w-full rounded-xl border border-slate-200 px-3 py-2 text-sm md:w-80"
                    />
                  </div>
                  <div className="overflow-hidden rounded-2xl border border-slate-200">
                  <table className="min-w-full divide-y divide-slate-200 text-sm">
                    <thead className="bg-slate-50">
                      <tr>
                        <th className="w-25 px-3 py-2 text-left font-medium text-slate-600">Waktu</th>
                        <th className="w-20 px-3 py-2 text-left font-medium text-slate-600">Aksi</th>
                        <th className="w-50 px-3 py-2 text-left font-medium text-slate-600">Modul</th>
                        <th className="px-3 py-2 text-left font-medium text-slate-600">Lokasi</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 bg-white">
                      {historyRows.length === 0 ? (
                        <tr>
                          <td colSpan={4} className="px-3 py-4 text-center text-slate-500">
                            Belum ada riwayat.
                          </td>
                        </tr>
                      ) : (
                        historyRows.map((item) => (
                          <tr key={item.id}>
                            <td className="px-3 py-2 text-slate-700">{formatDateTime(item.created_at)}</td>
                            <td className="px-3 py-2 text-slate-700">{item.action}</td>
                            <td className="px-3 py-2 text-slate-700">{item.module_code || "-"}</td>
                            <td className="px-3 py-2 text-slate-700">{item.path}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                  </div>
                  <div className="flex items-center justify-between rounded-xl border border-slate-200 px-3 py-2 text-sm text-slate-600">
                    <span>Halaman {historyPage} dari {historyPageTotal}</span>
                    <div className="flex gap-2">
                      <button type="button" disabled={historyPage <= 1} onClick={() => setHistoryPage((prev) => Math.max(1, prev - 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Prev</button>
                      <button type="button" disabled={historyPage >= historyPageTotal} onClick={() => setHistoryPage((prev) => Math.min(historyPageTotal, prev + 1))} className="rounded-lg border border-slate-200 px-3 py-1 disabled:opacity-50">Next</button>
                    </div>
                  </div>
                </div>
              ) : null}
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
